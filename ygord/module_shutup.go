// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
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

// ShutUpModule controls all the 'shut up', 'stop', 'sshhhh' commands.
type ShutUpModule struct{}

// Toggle determines whether the given message triggered this command.
func (module *ShutUpModule) Toggle(srv *Server, msg *Message) bool {
	body := strings.ToLower(msg.Body)
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

// PrivMsg is the message handler for user requests.
func (module *ShutUpModule) PrivMsg(srv *Server, msg *Message) {
	srv.SendToChannelMinions(msg.ReplyTo, "shutup")
	IRCPrivMsg(msg.ReplyTo, "ok...")
}

// Init registers all the commands for this module.
func (module ShutUpModule) Init() {
	RegisterCommand(Command{
		Name:            "shutup",
		ToggleFunction:  module.Toggle,
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
