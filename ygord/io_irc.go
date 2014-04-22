// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

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

// Send a message to a channel.
func IRCPrivMsg(channel, msg string) {
	lines := strings.Split(msg, "\n")
	for i := 0; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}
		IRCOutgoing <- fmt.Sprintf("PRIVMSG %s :%s", channel, lines[i])

		// Make test mode faster.
		if cfg.TestMode {
			time.Sleep(50 * time.Millisecond)
		} else {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func NewMessageFromPrivMsg(privmsg *ygor.PrivMsg) *ygor.Message {
	msg := ygor.NewMessage()

	if privmsg.Direct {
		msg.Type = ygor.MsgTypeIRCPrivate
	} else {
		// Channel messages have to be prefixes with the bot's nick.
		if !privmsg.Addressed {
			return nil
		}
		msg.Type = ygor.MsgTypeIRCChannel
	}

	msg.UserID = privmsg.Nick
	msg.Command = privmsg.Command
	msg.Body = privmsg.Body
	msg.ReplyTo = privmsg.ReplyTo
	msg.Args = privmsg.Args

	return msg
}

func NewMessageFromIRCSQS(sqsmsg *sqs.Message) *ygor.Message {
	msg := NewMessageFromIRCLine(strings.Trim(sqsmsg.Body, "\r\n"))
	if msg != nil {
		msg.SQSMessage = sqsmsg
	}
	return msg
}

// Create a new Message based on the raw IRC line.
func NewMessageFromIRCLine(line string) *ygor.Message {
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

	msg := NewMessageFromPrivMsg(privmsg)

	return msg
}

func ConvertIRCMessage(client *sqs.Client, sqsmsg *sqs.Message) error {
	msg := NewMessageFromIRCSQS(sqsmsg)

	if msg == nil {
		log.Printf("unhandled message: %s", sqsmsg.Body)
	} else {
		InputQueue <- msg
	}

	err := client.DeleteMessage(sqsmsg)
	if err != nil {
		return err
	}

	return nil
}

// Reads all input from the IRC incoming queue.
func ircIncomingHandler(client *sqs.Client) (<-chan error, error) {
	errch := make(chan error, 0)

	ch, sqserrch, err := sqschan.Incoming(client, cfg.IRCIncomingQueueName)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case sqsmsg := <-ch:
				err := ConvertIRCMessage(client, sqsmsg)
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
func ircOutgoingHandler(client *sqs.Client) (<-chan error, error) {
	ch, errch, err := sqschan.Outgoing(client, cfg.IRCOutgoingQueueName)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			ch <- <-IRCOutgoing
		}
	}()

	return errch, nil
}

// Start the incoming and outgoing handlers and multiplex their error channels.
func StartIRCAdapter(client *sqs.Client) (chan error, error) {
	errch := make(chan error, 0)

	incerrch, err := ircIncomingHandler(client)
	if err != nil {
		return nil, err
	}
	outerrch, err := ircOutgoingHandler(client)
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

	autojoin()

	return errch, nil
}

func IRCMessageHandler(msg *ygor.Message) {
	for _, cmd := range ygor.RegisteredCommands {
		if !cmd.IRCMessageMatches(msg) {
			continue
		}

		if cmd.PrivMsgFunction == nil {
			Debug("unhandled IRC message: " + msg.Body)
			continue
		}

		cmd.PrivMsgFunction(msg)
		break
	}
}

// Send an action message to a channel.
func privAction(channel, msg string) {
	IRCOutgoing <- fmt.Sprintf("PRIVMSG %s :\x01ACTION %s\x01", channel, msg)
}

func delayedPrivMsg(channel, msg string, waitTime time.Duration) {
	time.Sleep(waitTime)
	IRCPrivMsg(channel, msg)
}

func delayedPrivAction(channel, msg string, waitTime time.Duration) {
	time.Sleep(waitTime)
	privAction(channel, msg)
}

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
func autojoin() {
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
