// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"strings"
)

// var (
//     rePercentage = regexp.MustCompile(`^\d{1,3}%$`)
// )

// PVolumeModule is the module handling all the pvolume related commands.
type PVolumeModule struct{}

// PrivMsg is the message handler for user 'pvolume' requests.
func (module PVolumeModule) PrivMsg(srv *Server, msg *Message) {
	if len(msg.Args) != 1 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: pvolume percent")
		return
	}

	if !rePercentage.MatchString(msg.Args[0]) {
		srv.IRCPrivMsg(msg.ReplyTo, "error: bad input, must be absolute rounded percent value (e.g. 42%)")
		return
	}

	srv.SendToChannelMinions(msg.ReplyTo, "pvolume "+msg.Args[0])
}

// PrivMsgPlusPlus is the message handler for user 'pvolume++' requests, it
// increments the pvolume by 1dB.
func (module PVolumeModule) PrivMsgPlusPlus(srv *Server, msg *Message) {
	if len(msg.Args) != 0 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: pvolume++")
		return
	}
	srv.SendToChannelMinions(msg.ReplyTo, "pvolume 1dB+")
}

// PrivMsgMinusMinus is the message handler for user 'pvolume--' requests, it
// decrements the pvolume by 1dB.
func (module PVolumeModule) PrivMsgMinusMinus(srv *Server, msg *Message) {
	if len(msg.Args) != 0 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: pvolume--")
		return
	}
	srv.SendToChannelMinions(msg.ReplyTo, "pvolume 1dB-")
}

// MinionMsg is the message handler for all the minion responses for 'pvolume'
// requests.
func (module PVolumeModule) MinionMsg(srv *Server, msg *Message) {
	if msg.Args[0] != "ok" {
		minion, err := srv.Minions.GetByUserID(msg.UserID)
		if err != nil {
			log.Printf("pvolume: can't find minion for %s", msg.UserID)
			return
		}
		channels := srv.Config.GetChannelsByMinion(minion.Name)
		for _, channel := range channels {
			s := fmt.Sprintf("pvolume@%s: %s", minion.Name, strings.Join(msg.Args, " "))
			srv.IRCPrivMsg(channel, s)
		}
	}
}

// Init registers all the commands for this module.
func (module PVolumeModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:              "pvolume",
		PrivMsgFunction:   module.PrivMsg,
		MinionMsgFunction: module.MinionMsg,
		Addressed:         true,
		AllowPrivate:      false,
		AllowChannel:      true,
	})

	srv.RegisterCommand(Command{
		Name:            "pvolume++",
		PrivMsgFunction: module.PrivMsgPlusPlus,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	srv.RegisterCommand(Command{
		Name:            "pvolume--",
		PrivMsgFunction: module.PrivMsgMinusMinus,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
