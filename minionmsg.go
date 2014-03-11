// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package ygor

import (
	"regexp"
	"strings"
)

var (
	// Detect a MINIOMSG (minion communications).
	reMinionMsg = regexp.MustCompile(`^MINIONMSG (.*)`)
)

type MinionMsg struct {
	// The body of the message as received from the minion.
	Body string

	// Store the command and its arguments if relevant.
	Command string
	Args    []string
}

func NewMinionMsg(line string) *MinionMsg {
	msg := &MinionMsg{
		Body:      line,
	}

	tokens := strings.Split(msg.Body, " ")
	msg.Command = tokens[0]
	if len(tokens) > 1 {
		msg.Args = append(msg.Args, tokens[1:]...)
	}

	return msg
}
