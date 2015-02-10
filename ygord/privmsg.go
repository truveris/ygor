// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"regexp"
	"strings"
)

var (
	// Detect a PRIVMSG (most user communications).
	rePrivMsg = regexp.MustCompile(`^:([^!]+)![^@]+@[^\s]+\sPRIVMSG\s([^\s]+)\s:\s*(.*)\s*`)

	// Detect if we are addressed to.
	reAddressed = regexp.MustCompile(`^(\w+)[:,.]*\s*(.*)`)
)

// PrivMsg represents a message sent to the bot, be it through a channel or
// through a direct message.
type PrivMsg struct {
	// Who sent the message.
	Nick string

	// Where the message was sent to.
	Recipient string

	// Who to reply to.
	ReplyTo string

	// The whole message, minus the nickname if the message is addressed.
	Body string

	// Message received in the form of a /ME action.
	Action bool

	// The bot was addressed with its name as prefix (e.g. ygor: bla).
	Addressed bool

	// The bot was contacted directly instead of through a channel.
	Direct bool
}

// NewPrivMsg creates a new PrivMsg from the given IRC line and author.
func NewPrivMsg(line, nick string) *PrivMsg {
	tokens := rePrivMsg.FindStringSubmatch(line)
	if tokens == nil {
		return nil
	}

	msg := &PrivMsg{
		Nick:      tokens[1],
		Recipient: tokens[2],
		ReplyTo:   tokens[2],
		Body:      tokens[3],
		Addressed: false,
		Direct:    false,
	}

	if strings.HasPrefix(msg.Body, "\x01ACTION ") {
		msg.Body = msg.Body[8 : len(msg.Body)-1]
		msg.Action = true
	}

	// Message sent directly to the bot (not through a channel).
	if msg.Recipient == nick {
		msg.Addressed = true
		msg.Direct = true
		msg.ReplyTo = msg.Nick
	}

	// If the message is addressed (e.g. "ygor: hi"), remove the prefix
	// from the body and flag this message.
	tokens = reAddressed.FindStringSubmatch(msg.Body)
	if tokens != nil && tokens[1] == nick {
		msg.Addressed = true
		msg.Body = tokens[2]
	}

	return msg
}
