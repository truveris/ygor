// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"strings"
)

// XombreroModule is the module handling all the browser-related commands.
type XombreroModule struct{}

// PrivMsg is the message handler for 'xombrero' user requests.
func (module XombreroModule) PrivMsg(msg *Message) {
	if len(msg.Args) == 0 {
		IRCPrivMsg(msg.ReplyTo, "usage: xombrero [command [param ...]]")
		return
	}

	SendToChannelMinions(msg.ReplyTo, "xombrero "+strings.Join(msg.Args, " "))
}

// WebPrivMsg is the message handler for 'web' user requests.
func (module XombreroModule) WebPrivMsg(msg *Message) {
	if len(msg.Args) != 1 {
		IRCPrivMsg(msg.ReplyTo, "usage: web url")
		return
	}

	SendToChannelMinions(msg.ReplyTo, "xombrero open "+msg.Args[0])
}

// MinionMsg is the message handler for all messages coming from the minions.
func (module XombreroModule) MinionMsg(msg *Message) {
	if msg.Args[0] != "ok" {
		minion, err := Minions.GetByUserID(msg.UserID)
		if err != nil {
			Debug(fmt.Sprintf("xombrero: can't find minion for %s",
				msg.UserID))
			return
		}
		channels := GetChannelsByMinionName(minion.Name)
		for _, channel := range channels {
			s := fmt.Sprintf("xombrero@%s: %s", minion.Name, strings.Join(msg.Args, " "))
			IRCPrivMsg(channel, s)
		}
	}
}

// Init registers all the commands for this module.
func (module XombreroModule) Init() {
	RegisterCommand(Command{
		Name:              "xombrero",
		PrivMsgFunction:   module.PrivMsg,
		MinionMsgFunction: module.MinionMsg,
		Addressed:         true,
		AllowPrivate:      false,
		AllowChannel:      true,
	})

	RegisterCommand(Command{
		Name:            "web",
		PrivMsgFunction: module.WebPrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
