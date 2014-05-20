// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"strings"

	"github.com/truveris/ygor"
)

type SayModule struct{}

func (module SayModule) PrivMsg(msg *ygor.PrivMsg) {}

func SayCommand(msg *ygor.Message) {
	args := make([]string, 0)
	if len(msg.Args) > 3 {
		IRCPrivAction(msg.ReplyTo, "cropping to 3 words until someone works on ticket 7651")
		args = msg.Args[:3]
	} else {
		args = msg.Args
	}
	body := "say " + strings.Join(args, " ")
	SendToChannelMinions(msg.ReplyTo, body)
}

func (module SayModule) Init() {
	ygor.RegisterCommand(ygor.Command{
		Name:            "say",
		PrivMsgFunction: SayCommand,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
