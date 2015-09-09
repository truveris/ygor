// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"net/url"
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
	if err != nil || len(args) == 0 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: say [-v voice] sentence")
		return
	}

	if srv.Config.SaydURL == "" {
		srv.IRCPrivMsg(msg.ReplyTo, "error: SaydURL is not configured")
		return
	}

	src := srv.Config.SaydURL + cmd.Voice + ".mp3?" + url.QueryEscape(strings.Join(args, " "))
	mObj := new(MediaObj)
	err = mObj.SetSrc(src)
	if err != nil {
		srv.IRCPrivMsg(msg.ReplyTo, err.Error())
		return
	}

	mObj.Start = ""
	mObj.End = ""
	mObj.Muted = "false"

	json := "{" +
		"\"status\":\"media\"," +
		"\"track\":\"playTrack\"," +
		"\"loop\":false," +
		"\"mediaObjs\":[" +
		mObj.Serialize() +
		"]" +
		"}"

	// send command to minions
	srv.SendToChannelMinions(msg.ReplyTo,
		"play "+json)
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
