// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

var (
	rePercentage = regexp.MustCompile(`^[-+]?\d+%$`)
)

// VolumeModule is the module handling all the volume related commands.
type VolumeModule struct{}

// PrivMsg is the message handler for user 'volume' requests.
func (module VolumeModule) PrivMsg(srv *Server, msg *Message) {
	if len(msg.Args) != 1 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: volume percent")
		return
	}

	if !rePercentage.MatchString(msg.Args[0]) {
		srv.IRCPrivMsg(msg.ReplyTo, "error: bad input, must be percent")
		return
	}

	srv.SendToChannelMinions(msg.ReplyTo, "volume "+msg.Args[0])
}

// PrivMsgPlusPlus is the message handler for user 'volume++' requests, it
// increments the volume by 1dB.
func (module VolumeModule) PrivMsgPlusPlus(srv *Server, msg *Message) {
	if len(msg.Args) != 0 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: volume++")
		return
	}
	srv.SendToChannelMinions(msg.ReplyTo, "volume 1db+")
}

// PrivMsgMinusMinus is the message handler for user 'volume--' requests, it
// decrements the volume by 1dB.
func (module VolumeModule) PrivMsgMinusMinus(srv *Server, msg *Message) {
	if len(msg.Args) != 0 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: volume--")
		return
	}
	srv.SendToChannelMinions(msg.ReplyTo, "volume 1db-")
}

// MinionMsg is the message handler for all the minion responses for 'volume'
// requests.
func (module VolumeModule) MinionMsg(srv *Server, msg *Message) {
	if msg.Args[0] != "ok" {
		minion, err := srv.Minions.GetByUserID(msg.UserID)
		if err != nil {
			log.Printf("volume: can't find minion for %s", msg.UserID)
			return
		}
		channels := srv.Config.GetChannelsByMinion(minion.Name)
		for _, channel := range channels {
			s := fmt.Sprintf("volume@%s: %s", minion.Name, strings.Join(msg.Args, " "))
			srv.IRCPrivMsg(channel, s)
		}
	}
}

// Init registers all the commands for this module.
func (module VolumeModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:              "volume",
		PrivMsgFunction:   module.PrivMsg,
		MinionMsgFunction: module.MinionMsg,
		Addressed:         true,
		AllowPrivate:      false,
		AllowChannel:      true,
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
