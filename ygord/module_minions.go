// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This module allows for the registration and management of minions from IRC.
//

package main

import (
	"sort"
	"strings"
)

// MinionsModule controls the registration process for minions via the
// 'register' minion command.
type MinionsModule struct{}

// PrivMsg is the message handler for user 'minions' requests.
func (module *MinionsModule) PrivMsg(msg *Message) {
	var names []string

	minions, err := Minions.All()
	if err != nil {
		Debug("GetMinions error: " + err.Error())
		return
	}

	for _, minion := range minions {
		names = append(names, minion.Name)
	}
	sort.Strings(names)
	IRCPrivMsg(msg.ReplyTo, "currently registered: "+strings.Join(names, ", "))
}

// MinionMsg is the message handler for minions 'register' requests.
func (module *MinionsModule) MinionMsg(msg *Message) {
	if len(msg.Args) != 2 {
		Debug("register: error: invalid register command issued")
		return
	}

	name := msg.Args[0]
	queueURL := msg.Args[1]

	err := Minions.Register(name, queueURL, msg.UserID)
	if err != nil {
		Debug("register: error: " + err.Error())
		return
	}

	Minions.Save()
	err = SendToQueue(queueURL, "register success")
	if err != nil {
		Debug("register: error: " + err.Error())
	}
}

// Init registers all the commands for this module.
func (module *MinionsModule) Init() {
	RegisterCommand(Command{
		Name:              "register",
		MinionMsgFunction: module.MinionMsg,
		Addressed:         true,
		AllowPrivate:      true,
		AllowChannel:      true,
	})

	RegisterCommand(Command{
		Name:            "minions",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    true,
		AllowChannel:    true,
	})
}
