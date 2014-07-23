// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This module allows channel users to configure aliases themselves.

package main

import (
	"sort"
	"strings"

	"github.com/truveris/ygor"
)

type CommandsModule struct{}

func (module CommandsModule) PrivMsg(msg *ygor.PrivMsg) {}

func (module *CommandsModule) CommandsCmdFunc(msg *ygor.Message) {
	names := make([]string, 0)

	for name, cmd := range ygor.RegisteredCommands {
		// Attempt to only return user commands (skip minion commands).
		if cmd.PrivMsgFunction == nil {
			continue
		}

		names = append(names, name)
	}

	sort.Strings(names)

	found := strings.Join(names, ", ")

	IRCPrivMsg(msg.ReplyTo, found)
}

func (module *CommandsModule) Init() {
	ygor.RegisterCommand(ygor.Command{
		Name:            "commands",
		PrivMsgFunction: module.CommandsCmdFunc,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
