// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This module allows for the registration and management of minions from IRC.
//

package main

import (
	"sort"
	"strings"

	"github.com/truveris/ygor"
)

// Basic module.
type MinionsModule struct {
}

func (module *MinionsModule) PrivMsg(msg *ygor.Message) {}

func (module *MinionsModule) MinionsCmdFunc(msg *ygor.Message) {
	names := make([]string, 0)
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

// A minion is registering...
func (module *MinionsModule) RegisterMinionMsgFunc(msg *ygor.Message) {
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

func (module *MinionsModule) Init() {
	ygor.RegisterCommand(ygor.Command{
		Name:              "register",
		MinionMsgFunction: module.RegisterMinionMsgFunc,
		Addressed:         true,
		AllowPrivate:      true,
		AllowChannel:      true,
	})

	ygor.RegisterCommand(ygor.Command{
		Name:            "minions",
		PrivMsgFunction: module.MinionsCmdFunc,
		Addressed:       true,
		AllowPrivate:    true,
		AllowChannel:    true,
	})
}
