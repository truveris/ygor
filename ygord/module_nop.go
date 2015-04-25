// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// NopModule controls the 'nop' command.  This command does nothing at all.  It
// is used mostly to disable certain aliases from triggering ygor's "command
// not found" error.
type NopModule struct{}

// PrivMsg is the message handler for 'nop' requests.
func (module *NopModule) PrivMsg(srv *Server, msg *Message) {
}

// Init registers all the commands for this module.
func (module NopModule) Init() {
	RegisterCommand(Command{
		Name:            "nop",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
