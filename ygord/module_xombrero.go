// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"strings"

	"github.com/truveris/ygor"
)

type XombreroModule struct{}

// Send the xombrero command to the minions.
func (module XombreroModule) PrivMsg(msg *ygor.Message) {
	if len(msg.Args) == 0 {
		IRCPrivMsg(msg.ReplyTo, "usage: xombrero [command [param ...]]")
		return
	}

	SendToChannelMinions(msg.ReplyTo, "xombrero "+strings.Join(msg.Args, " "))
}

// Shortcut for xombrero open.
func (module XombreroModule) WebPrivMsg(msg *ygor.Message) {
	if len(msg.Args) != 1 {
		IRCPrivMsg(msg.ReplyTo, "usage: web url")
		return
	}

	SendToChannelMinions(msg.ReplyTo, "xombrero open "+msg.Args[0])
}

func (module XombreroModule) MinionMsg(msg *ygor.Message) {
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

func (module XombreroModule) Init() {
	ygor.RegisterCommand(ygor.Command{
		Name:              "xombrero",
		PrivMsgFunction:   module.PrivMsg,
		MinionMsgFunction: module.MinionMsg,
		Addressed:         true,
		AllowPrivate:      false,
		AllowChannel:      true,
	})

	ygor.RegisterCommand(ygor.Command{
		Name:            "web",
		PrivMsgFunction: module.WebPrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
