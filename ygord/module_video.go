// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// VideoModule controls the 'video' command.
type VideoModule struct {
	*Server
}

// PrivMsg is the message handler for user 'video' requests.
func (module *VideoModule) PrivMsg(srv *Server, msg *Message) {
	if len(msg.Args) != 1 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: video url")
		return
	}

	srv.SendToChannelMinions(msg.ReplyTo,
		"xombrero open http://truveris.github.io/fullscreen-video/?"+msg.Args[0])
}

// Init registers all the commands for this module.
func (module VideoModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:            "video",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
