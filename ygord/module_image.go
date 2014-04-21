// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"github.com/truveris/ygor"
)

type ImageModule struct{}

func (module ImageModule) PrivMsg(msg *ygor.Message) {}

func ImageFunc(msg *ygor.Message) {
	if len(msg.Args) != 1 {
		IRCPrivMsg(msg.ReplyTo, "usage: image url")
		return
	}

	SendToChannelMinions(msg.ReplyTo, "xombrero open http://truveris.github.io/fullscreen-image/?"+msg.Args[0])
}

func (module ImageModule) Init() {
	ygor.RegisterCommand(ygor.Command{
		Name:            "image",
		PrivMsgFunction: ImageFunc,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
