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
	body := "say " + strings.Join(msg.Args, " ")
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
