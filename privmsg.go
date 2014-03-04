// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package ygor

import (
	"regexp"
	"strings"
)

var (
	// Detect a PRIVMSG (most user communications).
	rePrivMsg = regexp.MustCompile(`^:([^!]+)![^@]+@[^\s]+\sPRIVMSG\s([^\s]+)\s:(.*)`)

	// Detect if we are addressed to.
	reAddressed = regexp.MustCompile(`^(\w+)[:,. ]+\s*(.*)`)
)

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

	// Store the command and its arguments if relevant.
	Command string
	Args    []string
}

func NewPrivMsg(line, currentNick string) *PrivMsg {
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
	if msg.Recipient == currentNick {
		msg.Addressed = true
		msg.Direct = true
		msg.ReplyTo = msg.Nick
	}

	// If the message is addressed (e.g. "ygor: hi"), remove the prefix
	// from the body and flag this message.
	tokens = reAddressed.FindStringSubmatch(msg.Body)
	if tokens != nil && tokens[1] == currentNick {
		msg.Addressed = true
		msg.Body = tokens[2]
	}

	tokens = strings.Split(msg.Body, " ")
	if len(tokens) > 0 {
		// Check if the first token is an alias.
		alias := GetAlias(tokens[0])
		if alias != nil {
			msg.Command, msg.Args = alias.SplitValue()
		} else {
			msg.Command = tokens[0]
		}

		// Did not find a matching alias, use the provided data.
		if msg.Command == "" {
			msg.Command = tokens[0]
		}

		if len(tokens) > 1 {
			msg.Args = append(msg.Args, tokens[1:]...)
		}
	}

	return msg
}
