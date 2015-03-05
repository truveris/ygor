// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// SkipModule controls the 'skip' command.
type SkipModule struct{}

// PrivMsg is the message handler for user requests.
func (module *SkipModule) PrivMsg(msg *Message) {
	SendToChannelMinions(msg.ReplyTo, "skip")
}

// Init registers all the commands for this module.
func (module SkipModule) Init() {
	RegisterCommand(Command{
		Name:            "skip",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
