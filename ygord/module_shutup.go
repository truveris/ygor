// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"regexp"
	"strings"

	"github.com/truveris/ygor"
)

var (
	reStop = regexp.MustCompile(`^st[aho]+p\b`)
	reShhh = regexp.MustCompile(`^s+[sh]+\b`)
)

type ShutUpModule struct{}

func (module ShutUpModule) PrivMsg(msg *ygor.PrivMsg) {}

func isShutUpRequest(msg *ygor.Message) bool {
	body := strings.ToLower(msg.Body)
	println(body)
	if reStop.MatchString(body) {
		return true
	}
	if reShhh.MatchString(body) {
		return true
	}
	if strings.HasPrefix(body, "shut up") {
		return true
	}
	return false
}

func ShutUpCommand(msg *ygor.Message) {
	SendToChannelMinions(msg.ReplyTo, "shutup")
	IRCPrivMsg(msg.ReplyTo, "ok...")
}

func (module ShutUpModule) Init() {
	ygor.RegisterCommand(ygor.Command{
		Name:            "shutup",
		ToggleFunction:  isShutUpRequest,
		PrivMsgFunction: ShutUpCommand,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
