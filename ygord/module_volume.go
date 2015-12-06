// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"regexp"
)

var (
	rePercentage = regexp.MustCompile(`^\d{1,3}%$`)
)

// VolumeModule is the module handling all the volume related commands.
type VolumeModule struct{}

// PrivMsg is the message handler for user 'volume' requests.
func (module VolumeModule) PrivMsg(srv *Server, msg *IRCInputMessage) {
	if len(msg.Args) != 1 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: volume percent")
		return
	}

	if !rePercentage.MatchString(msg.Args[0]) {
		srv.IRCPrivMsg(msg.ReplyTo, "error: bad input, must be absolute rounded percent value (e.g. 42%)")
		return
	}

	srv.SendToChannelMinions(msg.ReplyTo, "volume "+msg.Args[0])
}

// PrivMsgPlusPlus is the message handler for user 'volume++' requests, it
// increments the volume by 1dB.
func (module VolumeModule) PrivMsgPlusPlus(srv *Server, msg *IRCInputMessage) {
	if len(msg.Args) != 0 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: volume++")
		return
	}
	srv.SendToChannelMinions(msg.ReplyTo, "volume 1dB+")
}

// PrivMsgMinusMinus is the message handler for user 'volume--' requests, it
// decrements the volume by 1dB.
func (module VolumeModule) PrivMsgMinusMinus(srv *Server, msg *IRCInputMessage) {
	if len(msg.Args) != 0 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: volume--")
		return
	}
	srv.SendToChannelMinions(msg.ReplyTo, "volume 1dB-")
}

// Init registers all the commands for this module.
func (module VolumeModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:            "volume",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	srv.RegisterCommand(Command{
		Name:            "volume++",
		PrivMsgFunction: module.PrivMsgPlusPlus,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	srv.RegisterCommand(Command{
		Name:            "volume--",
		PrivMsgFunction: module.PrivMsgMinusMinus,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
