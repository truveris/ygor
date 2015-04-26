// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// TODO: Make this module wait for 3 request from three users within 15 minutes
// to execute the reboot.
//

package main

// RebootModule controls the 'reboot' function.
type RebootModule struct{}

// PrivMsg is the message handler for user requests.
func (module *RebootModule) PrivMsg(srv *Server, msg *Message) {
	srv.SendToChannelMinions(msg.ReplyTo, "reboot")
	srv.IRCPrivMsg(msg.ReplyTo, "attempting to reboot "+msg.ReplyTo+" minions...")
}

// Init registers all the commands for this module.
func (module RebootModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:            "reboot",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
