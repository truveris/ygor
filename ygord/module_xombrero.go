// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	//	"fmt"
	"strings"

	"github.com/truveris/ygor"
)

type XombreroModule struct{}

func (module XombreroModule) PrivMsg(msg *ygor.Message) {
	if len(msg.Args) == 0 {
		IRCPrivMsg(msg.ReplyTo, "usage: xombrero [command [param ...]]")
		return
	}

	SendToChannelMinions(msg.ReplyTo, "xombrero "+strings.Join(msg.Args, " "))
}

// func (module XombreroModule) MinionMsg(msg *ygor.MinionMsg) {
// 	if msg.Args[0] == "ok" {
// 		channels := GetChannelsByMinionName(msg.Name)
// 		for _, channel := range channels {
// 			s := fmt.Sprintf("%s: %s", msg.Name, okWord())
// 			IRCPrivMsg(channel, s)
// 		}
// 	}
// }

func (module XombreroModule) Init() {
	ygor.RegisterCommand(ygor.Command{
		Name:            "xombrero",
		PrivMsgFunction: module.PrivMsg,
		// MinionMsgFunction: module.MinionMsg,
		Addressed:    true,
		AllowPrivate: false,
		AllowChannel: true,
	})
}
