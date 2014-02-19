// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

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
	Nick string
	Channel string
	Body string
	IsAction bool
	IsAddressed bool
	Command string
	Args []string
}

func NewPrivMsg(line string) *PrivMsg {
	tokens := rePrivMsg.FindStringSubmatch(line)
	if tokens == nil {
		return nil
	}

	msg := &PrivMsg{
		Nick: tokens[1],
		Channel: tokens[2],
		Body: tokens[3],
	}

	if strings.HasPrefix(msg.Body, "\x01ACTION ") {
		msg.Body = msg.Body[8 : len(msg.Body)-1]
		msg.IsAction = true
	}

	// If the message is addressed (e.g. "ygor: hi"), remove the prefix
	// from the body and flag this message.
	msg.IsAddressed = false
	tokens = reAddressed.FindStringSubmatch(msg.Body)
	if tokens != nil && tokens[1] == cmd.Nickname {
		msg.IsAddressed = true
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
