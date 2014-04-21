// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"log"
	"strings"

	"github.com/truveris/sqs"
	"github.com/truveris/sqs/sqschan"
	"github.com/truveris/ygor"
)

var (
	IRCOutgoing = make(chan string, 0)
	IRCNickname string
)

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
		inputQueue <- msg
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
func startIRCHandlers(client *sqs.Client) (chan error, error) {
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

	return errch, nil
}
