// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"strings"

	"github.com/truveris/ygor"
	"github.com/jessevdk/go-flags"
)

type SayCmd struct {
	Voice string `short:"v" description:"Voice" default:"bruce"`
}

type SayModule struct{}

func (module SayModule) PrivMsg(msg *ygor.PrivMsg) {}

func SayCommand(msg *ygor.Message) {
	cmd := SayCmd{}

	flagParser := flags.NewParser(&cmd, flags.PassDoubleDash)
	args, err := flagParser.ParseArgs(msg.Args)
	if err != nil {
		IRCPrivMsg(msg.ReplyTo, "usage: say [-v voice] sentence")
		return
	}

	body := "say -v " + cmd.Voice + " " + strings.Join(args, " ")
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
