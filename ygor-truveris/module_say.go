// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"strings"
)

type SayModule struct{}

func (module SayModule) PrivMsg(msg *PrivMsg) {}

func SayCommand(msg *PrivMsg) {
	body := "say " + strings.Join(msg.Args, " ")
	SendToMinion(msg.ReplyTo, body)
}

func (module SayModule) Init() {
	RegisterCommand(Command{
		Name:         "say",
		Function:     SayCommand,
		Addressed:    true,
		AllowDirect:  false,
		AllowChannel: true,
	})
}
