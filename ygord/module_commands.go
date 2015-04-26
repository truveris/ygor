// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This module allows channel users to configure aliases themselves.

package main

import (
	"sort"
	"strings"
)

// CommandsModule controls the 'commands' command which lists all the known
// commands publicly.
type CommandsModule struct{}

// PrivMsg is the message handler for user 'commands' requests.
func (module *CommandsModule) PrivMsg(srv *Server, msg *Message) {
	var names []string

	for name, cmd := range srv.RegisteredCommands {
		// Attempt to only return user commands (skip minion commands).
		if cmd.PrivMsgFunction == nil {
			continue
		}

		names = append(names, name)
	}

	sort.Strings(names)

	found := strings.Join(names, ", ")

	srv.IRCPrivMsg(msg.ReplyTo, found)
}

// Init registers all the commands for this module.
func (module *CommandsModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:            "commands",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
