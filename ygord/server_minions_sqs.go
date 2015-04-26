// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// The minion-sqs adapter allows ygord to send and receive data to the minions
// via SQS.
//

package main

import (
	"log"
	"strings"

	"github.com/truveris/sqs"
	"github.com/truveris/sqs/sqschan"
)

// StartMinionAdapter is the entry point for this IO adapter. It reads from the
// main ygord queue and assume all the incoming messages are minion feedbacks.
func (srv *Server) StartMinionAdapter(client *sqs.Client, queueName string) (<-chan error, error) {
	errch := make(chan error, 0)

	ch, sqserrch, err := sqschan.Incoming(client, queueName)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case sqsmsg := <-ch:
				msg := NewMessageFromMinionSQS(sqsmsg)
				srv.InputQueue <- msg
				err := client.DeleteMessage(sqsmsg)
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

// MinionMessageHandler is used from main() when receiving data on the MinionsIncoming.
func (srv *Server) MinionMessageHandler(msg *Message) {
	for _, cmd := range srv.RegisteredCommands {
		if !cmd.MinionMessageMatches(msg) {
			continue
		}

		if cmd.MinionMsgFunction == nil {
			log.Printf("unhandled minion message: %s", msg.Body)
			continue
		}

		cmd.MinionMsgFunction(srv, msg)
		break
	}
}

// NewMessageFromMinionLine converts a raw minion line into an ygor message.
func NewMessageFromMinionLine(line string) *Message {
	msg := NewMessage()
	msg.Type = MsgTypeMinion
	msg.Body = line

	args := strings.Split(line, " ")
	msg.Command = args[0]
	msg.Args = args[1:]

	return msg
}

// NewMessageFromMinionSQS converts an SQS message into an ygor message.
func NewMessageFromMinionSQS(sqsmsg *sqs.Message) *Message {
	msg := NewMessageFromMinionLine(sqs.SQSDecode(sqsmsg.Body))
	msg.UserID = sqsmsg.SenderID
	return msg
}
