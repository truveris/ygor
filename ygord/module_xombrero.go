// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"strings"
)

// XombreroModule is the module handling all the browser-related commands.
type XombreroModule struct{}

// PrivMsg is the message handler for 'xombrero' user requests.
func (module XombreroModule) PrivMsg(srv *Server, msg *Message) {
	if len(msg.Args) == 0 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: xombrero [command [param ...]]")
		return
	}

	srv.SendToChannelMinions(msg.ReplyTo, "xombrero "+strings.Join(msg.Args, " "))
}

// WebPrivMsg is the message handler for 'web' user requests.
func (module XombreroModule) WebPrivMsg(srv *Server, msg *Message) {
	if len(msg.Args) != 1 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: web url")
		return
	}

	srv.SendToChannelMinions(msg.ReplyTo, "xombrero open "+msg.Args[0])
}

// MinionMsg is the message handler for all messages coming from the minions.
func (module XombreroModule) MinionMsg(srv *Server, msg *Message) {
	if msg.Args[0] != "ok" {
		minion, err := srv.Minions.GetByUserID(msg.UserID)
		if err != nil {
			log.Printf("xombrero: can't find minion for %s", msg.UserID)
			return
		}
		channels := srv.Config.GetChannelsByMinion(minion.Name)
		for _, channel := range channels {
			s := fmt.Sprintf("xombrero@%s: %s", minion.Name, strings.Join(msg.Args, " "))
			srv.IRCPrivMsg(channel, s)
		}
	}
}

// Init registers all the commands for this module.
func (module XombreroModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:              "xombrero",
		PrivMsgFunction:   module.PrivMsg,
		MinionMsgFunction: module.MinionMsg,
		Addressed:         true,
		AllowPrivate:      false,
		AllowChannel:      true,
	})

	srv.RegisterCommand(Command{
		Name:            "web",
		PrivMsgFunction: module.WebPrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
