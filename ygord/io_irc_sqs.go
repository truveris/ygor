// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// The io_irc_sqs portion of the ygor code base defines the adapter used to
// receive and send data to IRC via SQS (assuming irc-sqs-gateway is setup).
//
// The message in this adapter is roughly converted as such:
//
//     sqs.Message -> ygor.PrivMsg -> ygor.Message
//

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/truveris/sqs"
	"github.com/truveris/sqs/sqschan"
	"github.com/truveris/ygor"
)

var (
	IRCOutgoing = make(chan string, 0)
	IRCNickname string
)

// Create a new ygor message from a parsed PRIVMSG.
func NewMessagesFromPrivMsg(privmsg *ygor.PrivMsg) []*ygor.Message {
	msgs := make([]*ygor.Message, 0)
	bodies := strings.Split(privmsg.Body, ";")

	for _, body := range bodies {
		msg := ygor.NewMessage()

		if privmsg.Direct {
			msg.Type = ygor.MsgTypeIRCPrivate
		} else {
			// Channel messages have to be prefixes with the bot's nick.
			if !privmsg.Addressed && !privmsg.Direct {
				return nil
			}
			msg.Type = ygor.MsgTypeIRCChannel
		}

		msg.UserID = privmsg.Nick
		msg.ReplyTo = privmsg.ReplyTo

		// Resolve any aliases.
		body, err := Aliases.Resolve(strings.Trim(body, " \r\n\t"))
		if err != nil {
			IRCPrivMsg(msg.ReplyTo, "failed to resolve aliases: " +
				err.Error())
			continue
		}

		msg.Body = body

		tokens := strings.Split(msg.Body, " ")
		if len(tokens) > 0 {
			msg.Command = tokens[0]

			if len(tokens) > 1 {
				msg.Args = append(msg.Args, tokens[1:]...)
			}
		}

		msgs = append(msgs, msg)
	}

	return msgs
}

// Create a new ygor message from an SQS message.
func NewMessagesFromIRCSQS(sqsmsg *sqs.Message) []*ygor.Message {
	msgs := NewMessagesFromIRCLine(strings.Trim(sqsmsg.Body, "\r\n"))
	for _, msg := range msgs {
		if msg != nil {
			msg.SQSMessage = sqsmsg
		}
	}
	return msgs
}

// Create a new Message based on the raw IRC line.
func NewMessagesFromIRCLine(line string) []*ygor.Message {
	privmsg := ygor.NewPrivMsg(line, cfg.IRCNickname)
	if privmsg == nil {
		// Not a PRIVMSG.
		return nil
	}

	// Check if we should ignore this message.
	for _, ignore := range cfg.Ignore {
		if ignore == privmsg.Nick {
			return nil
		}
	}

	return NewMessagesFromPrivMsg(privmsg)
}

// Convert an SQS message to an ygor Message and feed it to the InputQueue if
// it is a valid IRC message. The SQS message is then deleted.
func ReceiveSQSMessageForIRC(client *sqs.Client, sqsmsg *sqs.Message) error {
	msgs := NewMessagesFromIRCSQS(sqsmsg)

	for _, msg := range msgs {
		if msg == nil {
			log.Printf("unhandled message in line: %s", sqsmsg.Body)
		} else {
			InputQueue <- msg
		}
	}

	err := client.DeleteMessage(sqsmsg)
	if err != nil {
		return err
	}

	return nil
}

// Reads all input from the IRC incoming queue passing errors to the error
// channel.
func StartIRCIncomingQueueReader(client *sqs.Client) (<-chan error, error) {
	errch := make(chan error, 0)

	ch, sqserrch, err := sqschan.Incoming(client, cfg.IRCIncomingQueueName)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case sqsmsg := <-ch:
				err := ReceiveSQSMessageForIRC(client, sqsmsg)
				if err != nil {
					errch <- err
				}
			case err := <-sqserrch:
				errch <- err
			}
		}
	}()

	return errch, nil
}

// Write all the messages from the outgoing channel.
func StartIRCOutgoingQueueWriter(client *sqs.Client) (<-chan error, error) {
	ch, errch, err := sqschan.Outgoing(client, cfg.IRCOutgoingQueueName)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			ch <- sqs.SQSEncode(<-IRCOutgoing)
		}
	}()

	return errch, nil
}

// Start the incoming and outgoing handlers and multiplex their error channels.
func StartIRCAdapter(client *sqs.Client) (chan error, error) {
	errch := make(chan error, 0)

	incerrch, err := StartIRCIncomingQueueReader(client)
	if err != nil {
		return nil, err
	}
	outerrch, err := StartIRCOutgoingQueueWriter(client)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case err := <-incerrch:
				errch <- err
			case err := <-outerrch:
				errch <- err
			}
		}
	}()

	AutoJoin()

	return errch, nil
}

func IRCMessageHandler(msg *ygor.Message) {
	for _, cmd := range ygor.RegisteredCommands {
		if !cmd.IRCMessageMatches(msg) {
			continue
		}

		if cmd.PrivMsgFunction == nil {
			log.Printf("misconfigured command: %s (no PrivMsg)",
				cmd.Name)
			continue
		}

		cmd.PrivMsgFunction(msg)
		return
	}

	// If we got that far, we didn't find a command.
	IRCPrivMsg(msg.ReplyTo, "command not found: "+msg.Command)
}

//
// All the functions below are used as wrappers to send data to the IRC server.
// They are convenience functions for ygor to speak.
//

// Send a message to a channel. Construct a PRIVMSG and send the raw client
// line to the server (via SQS).
func IRCPrivMsg(channel, msg string) {
	lines := strings.Split(msg, "\n")
	for i := 0; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}
		IRCOutgoing <- fmt.Sprintf("PRIVMSG %s :%s", channel, lines[i])
	}
}

// Send an action message to a channel. This function is the equivalent of
// using the /ME command in a normal IRC client.
func IRCPrivAction(channel, msg string) {
	IRCOutgoing <- fmt.Sprintf("PRIVMSG %s :\x01ACTION %s\x01", channel, msg)
}

// Send a /JOIN command to the server.
func JoinChannel(channel string) {
	IRCOutgoing <- "JOIN " + channel
}

// Send the message to the configured debug channel if any.
func Debug(msg string) {
	if cfg.AdminChannel == "" {
		return
	}
	IRCPrivMsg(cfg.AdminChannel, msg)
}

// Auto-join all the configured channels.
func AutoJoin() {
	// Make test mode faster.
	if cfg.TestMode {
		time.Sleep(50 * time.Millisecond)
	} else {
		time.Sleep(500 * time.Millisecond)
	}

	for _, c := range cfg.GetAutoJoinChannels() {
		JoinChannel(c)
	}
}
