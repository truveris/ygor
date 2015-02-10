// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
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
)

var (
	// IRCOutgoing is the queue of all outgoing IRC messages.  All strings
	// going into this channel should formatted exactly as if they were
	// sent directly to the IRC server.
	IRCOutgoing = make(chan string, 0)
)

func expandSentence(words []string) ([][]string, error) {
	sentences := make([][]string, len(words))

	// Resolve any alias found as first word.
	expanded, err := Aliases.Resolve(words[0])
	if err != nil {
		return nil, err
	}

	sentences, err = LexerSplit(expanded)
	if err != nil {
		return nil, err
	}

	last := len(sentences) - 1
	sentences[last] = append(sentences[last], words[1:]...)

	return sentences, nil
}

// Expand sentences through aliases.
func expandSentences(ss [][]string) ([][]string, error) {
	sentences := make([][]string, len(ss))

	for _, words := range ss {
		if len(words) == 0 {
			continue
		}

		newsentences, err := expandSentence(words)
		if err != nil {
			return nil, err
		}
		sentences = append(sentences, newsentences...)
	}

	return sentences, nil
}

// NewMessagesFromBody creates a new ygor message from a plain string.
func NewMessagesFromBody(body string) ([]*Message, error) {
	var msgs []*Message

	sentences, err := LexerSplit(body)
	if err != nil {
		return nil, err
	}

	// TODO: make that recursive.
	for i := 0; i < 3; i++ {
		sentences, err = expandSentences(sentences)
		if err != nil {
			return nil, err
		}
	}

	for _, words := range sentences {
		if len(words) == 0 {
			continue
		}

		msg := NewMessage()

		msg.Body = strings.Join(words, " ")
		msg.Command = words[0]

		if len(words) > 1 {
			msg.Args = append(msg.Args, words[1:]...)
		}

		msgs = append(msgs, msg)
	}

	return msgs, nil
}

// NewMessagesFromPrivMsg create a new ygor message from a parsed PRIVMSG.
func NewMessagesFromPrivMsg(privmsg *PrivMsg) []*Message {
	msgs, err := NewMessagesFromBody(privmsg.Body)
	if err != nil {
		if privmsg.Addressed {
			IRCPrivMsg(privmsg.ReplyTo, "lexer/expand error: "+
				err.Error())
		}
		return nil
	}

	for _, msg := range msgs {
		if privmsg.Direct {
			msg.Type = MsgTypeIRCPrivate
		} else {
			// Channel messages have to be prefixes with the bot's nick.
			if !privmsg.Addressed && !privmsg.Direct {
				return nil
			}
			msg.Type = MsgTypeIRCChannel
		}

		msg.UserID = privmsg.Nick
		msg.ReplyTo = privmsg.ReplyTo
	}

	return msgs
}

// NewMessagesFromIRCSQS creates a new message from an SQS message.
func NewMessagesFromIRCSQS(sqsmsg *sqs.Message) []*Message {
	msgs := NewMessagesFromIRCLine(strings.Trim(sqsmsg.Body, "\r\n"))
	for _, msg := range msgs {
		if msg != nil {
			msg.SQSMessage = sqsmsg
		}
	}
	return msgs
}

// NewMessagesFromIRCLine creates a new Message based on the raw IRC line.
func NewMessagesFromIRCLine(line string) []*Message {
	privmsg := NewPrivMsg(line, cfg.IRCNickname)
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

// ReceiveSQSMessageForIRC converts an SQS message to an ygor Message and feed
// it to the InputQueue if it is a valid IRC message.  The SQS message is then
// deleted.
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

// StartIRCIncomingQueueReader reads all input from the IRC incoming queue
// passing errors to the error channel.
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

// StartIRCOutgoingQueueWriter writes all the messages from the outgoing channel.
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

// StartIRCAdapter boots the incoming and outgoing handlers and multiplex their
// error channels.
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

// IRCMessageHandler loops through the command registry to find a matching
// command and executes it.
func IRCMessageHandler(msg *Message) {
	for _, cmd := range RegisteredCommands {
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

// IRCPrivMsg sends a message to a channel. Construct a PRIVMSG and send the
// raw client line to the server (via SQS).
func IRCPrivMsg(channel, msg string) {
	lines := strings.Split(msg, "\n")
	for i := 0; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}
		IRCOutgoing <- fmt.Sprintf("PRIVMSG %s :%s", channel, lines[i])
	}
}

// IRCPrivAction sends an action message to a channel.  This function is the
// equivalent of using the /ME command in a normal IRC client.
func IRCPrivAction(channel, msg string) {
	IRCOutgoing <- fmt.Sprintf("PRIVMSG %s :\x01ACTION %s\x01", channel, msg)
}

// JoinChannel sends a /JOIN command to the server.
func JoinChannel(channel string) {
	IRCOutgoing <- "JOIN " + channel
}

// Debug sends the message to the configured debug channel if any.
func Debug(msg string) {
	if cfg.AdminChannel == "" {
		return
	}
	IRCPrivMsg(cfg.AdminChannel, msg)
}

// AutoJoin all the configured channels.
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
