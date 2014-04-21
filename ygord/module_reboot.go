// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// TODO: Make this module wait for 3 request within 15 minutes to execute the
// reboot.
//

package main

import (
	"github.com/truveris/ygor"
)

type RebootModule struct{}

func (module RebootModule) PrivMsg(msg *ygor.PrivMsg) {}

func RebootFunc(msg *ygor.Message) {
	SendToChannelMinions(msg.ReplyTo, "reboot")
	IRCPrivMsg(msg.ReplyTo, "attempting to reboot "+msg.ReplyTo+" minions...")
}

func (module RebootModule) Init() {
	ygor.RegisterCommand(ygor.Command{
		Name:            "reboot",
		PrivMsgFunction: RebootFunc,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
