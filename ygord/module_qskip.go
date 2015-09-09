// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// QSkipModule controls the 'qskip' command.
type QSkipModule struct{}

// PrivMsg is the message handler for user requests.
func (module *QSkipModule) PrivMsg(srv *Server, msg *Message) {
	srv.SendToChannelMinions(msg.ReplyTo, "qskip")
}

// Init registers all the commands for this module.
func (module QSkipModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:            "qskip",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
