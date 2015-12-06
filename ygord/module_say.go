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
func (module *SayModule) PrivMsg(srv *Server, msg *IRCInputMessage) {
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

	sayURL := srv.Config.SaydURL + cmd.Voice + ".mp3?" + url.QueryEscape(strings.Join(args, " "))
	mediaItem := make(map[string]string)
	mediaItem["url"] = sayURL

	media, parseMObjErr := NewMedia(srv, mediaItem, "playTrack", false, false,
		[]string{})
	if parseMObjErr != nil {
		srv.IRCPrivMsg(msg.ReplyTo, parseMObjErr.Error())
		return
	}

	// Override the formatted Src to be the original sayURL, because, in this
	// case, the query string is needed.
	media.Src = sayURL

	// Send the command to the connected minions, as though it were the play
	// command.
	srv.SendToChannelMinions(msg.ReplyTo,
		"play "+media.Serialize())
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
