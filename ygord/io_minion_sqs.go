// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// The minion-sqs adapter allows ygord to send and receive data to the minions
// via SQS.
//

package main

import (
	"fmt"
	"strings"

	"github.com/truveris/sqs"
	"github.com/truveris/sqs/sqschan"
)

// StartMinionAdapter is the entry point for this IO adapter. It reads from the
// main ygord queue and assume all the incoming messages are minion feedbacks.
func StartMinionAdapter(client *sqs.Client) (<-chan error, error) {
	errch := make(chan error, 0)

	ch, sqserrch, err := sqschan.Incoming(client, cfg.QueueName)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case sqsmsg := <-ch:
				msg := NewMessageFromMinionSQS(sqsmsg)
				InputQueue <- msg
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

// SendToChannelMinions sends a message to all the minions of the given
// channel.
func SendToChannelMinions(channel, msg string) {
	client, err := sqs.NewClient(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey,
		cfg.AWSRegionCode)
	if err != nil {
		Debug("error: " + err.Error())
		return
	}

	channelCfg, exists := cfg.Channels[channel]
	if !exists {
		Debug("error: " + channel + " has no queue(s) configured")
		return
	}

	urls, err := channelCfg.GetQueueURLs()
	if err != nil {
		Debug("error: unable to load queue URLs, " + err.Error())
		return
	}

	// Send the same exact data to all this channel's minion.
	for _, url := range urls {
		if cfg.TestMode {
			fmt.Printf("[SQS-SendToMinion] %s %s\n", url, msg)
			continue
		}

		err := client.SendMessage(url, sqs.SQSEncode(msg))
		if err != nil {
			Debug("error sending to minion: " + err.Error())
			continue
		}
	}
}

// SendToQueue sends a message to our friendly minion via its SQS queue.
func SendToQueue(queueURL, msg string) error {
	if cfg.TestMode {
		fmt.Printf("[SQS-SendToMinion] %s %s\n", queueURL, msg)
		return nil
	}

	client, err := sqs.NewClient(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey,
		cfg.AWSRegionCode)
	if err != nil {
		return err
	}

	err = client.SendMessage(queueURL, msg)
	if err != nil {
		Debug("error sending to minion: " + err.Error())
	}

	return nil
}

// MinionMessageHandler is used from main() when receiving data on the InputQueue.
func MinionMessageHandler(msg *Message) {
	for _, cmd := range RegisteredCommands {
		if !cmd.MinionMessageMatches(msg) {
			continue
		}

		if cmd.MinionMsgFunction == nil {
			Debug("unhandled minion message: " + msg.Body)
			continue
		}

		cmd.MinionMsgFunction(msg)
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
