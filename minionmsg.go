// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// Defines all the tools to handle the MINIONMSG messages coming from the
// minions.
//
// TODO: We need to reject unauthenticated messages (wrong user id).
//

package ygor

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// Detect a MINIOMSG (minion communications).
	reMinionMsg = regexp.MustCompile(`^([^\s]+) MINIONMSG (.*)`)
)

type MinionMsg struct {
	// Name of the minion sending this message. This value is populated
	// from the registry, only if we can find a matching entry using the
	// UserID.
	Name string

	// User sending the message. This value is authenticated.
	UserID string

	// The body of the message as received from the minion.
	Body string

	// Store the command and its arguments if relevant.
	Command string
	Args    []string
}

func NewMinionMsg(line string) (*MinionMsg, error) {
	tokens := reMinionMsg.FindStringSubmatch(line)
	if tokens == nil {
		return nil, errors.New("not a proper MINIONMSG")
	}

	msg := &MinionMsg{
		UserID: tokens[1],
		Body: tokens[2],
	}

	tokens = strings.Split(msg.Body, " ")
	msg.Command = tokens[0]
	if len(tokens) > 1 {
		msg.Args = append(msg.Args, tokens[1:]...)
	}

	// At the time of registration, we cannot guarantee that a name is
	// available, skip the minion lookup entirely.
	if msg.Command == "register" {
		return msg, nil
	}

	minion, err := GetMinionByUserID(msg.UserID)
	if err != nil {
		return nil, err
	}

	msg.Name = minion.Name

	return msg, nil
}
