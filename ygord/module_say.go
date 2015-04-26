// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"strings"

	"github.com/jessevdk/go-flags"
)

// SayCmdLine is the schema for the command-line parser of the 'say' command.
type SayCmdLine struct {
	Voice string `short:"v" description:"Voice" default:"bruce"`
}

// SayModule controls all the 'say' commands.
type SayModule struct{}

// PrivMsg is the message handler for user requests.
func (module *SayModule) PrivMsg(srv *Server, msg *Message) {
	cmd := SayCmdLine{}

	flagParser := flags.NewParser(&cmd, flags.PassDoubleDash)
	args, err := flagParser.ParseArgs(msg.Args)
	if err != nil {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: say [-v voice] sentence")
		return
	}

	body := "say -v " + cmd.Voice + " " + strings.Join(args, " ")
	srv.SendToChannelMinions(msg.ReplyTo, body)
}

// Init registers all the commands for this module.
func (module SayModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:            "say",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
