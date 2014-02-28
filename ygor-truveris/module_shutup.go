// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"regexp"
	"strings"
)

var (
	reStop = regexp.MustCompile(`^st[aho]+p\b`)
	reShhh = regexp.MustCompile(`^s+[sh]+\b`)
)

type ShutUpModule struct{}

func (module ShutUpModule) PrivMsg(msg *PrivMsg) {}

func isShutUpRequest(msg *PrivMsg) bool {
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

func ShutUpCommand(msg *PrivMsg) {
	SendToMinion(msg.ReplyTo, "shutup")
	privMsg(msg.ReplyTo, "ok...")
}

func (module ShutUpModule) Init() {
	RegisterCommand(Command{
		Name:           "shutup",
		ToggleFunction: isShutUpRequest,
		Function:       ShutUpCommand,
		Addressed:      true,
		AllowDirect:    false,
		AllowChannel:   true,
	})
}
