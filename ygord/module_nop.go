// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// Do nothing. Because reasons.

package main

import (
	"github.com/truveris/ygor"
)

type NopModule struct{}

func (module NopModule) PrivMsg(msg *ygor.PrivMsg) {}

func NopFunc(msg *ygor.Message) {
}

func (module NopModule) Init() {
	ygor.RegisterCommand(ygor.Command{
		Name:            "nop",
		PrivMsgFunction: NopFunc,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
