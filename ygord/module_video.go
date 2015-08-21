// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// VideoModule controls the 'video' command.
type VideoModule struct{}

// PrivMsg is the message handler for user 'video' requests.
func (module *VideoModule) PrivMsg(msg *Message) {
	if len(msg.Args) != 1 {
		IRCPrivMsg(msg.ReplyTo, "usage: video url")
		return
	}

	SendToChannelMinions(msg.ReplyTo, "xombrero open http://truveris.github.io/fullscreen-video/?"+msg.Args[0])
}

// Init registers all the commands for this module.
func (module VideoModule) Init() {
	RegisterCommand(Command{
		Name:            "video",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
