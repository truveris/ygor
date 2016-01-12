// Copyright 2016, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// ScreensaverModule controls the 'image' command.
type ScreensaverModule struct {
	*Server
}

// PrivMsg is the message handler for user 'image' requests.
func (module *ScreensaverModule) PrivMsg(srv *Server, msg *InputMessage) {
	srv.Reply(msg, "error: not implemented yet")
}

// Init registers all the commands for this module.
func (module ScreensaverModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:            "screensaver",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
