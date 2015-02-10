// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// ImageModule controls the 'image' command.
type ImageModule struct{}

// PrivMsg is the message handler for user 'image' requests.
func (module *ImageModule) PrivMsg(msg *Message) {
	if len(msg.Args) != 1 {
		IRCPrivMsg(msg.ReplyTo, "usage: image url")
		return
	}

	SendToChannelMinions(msg.ReplyTo, "xombrero open http://truveris.github.io/fullscreen-image/?"+msg.Args[0])
}

// Init registers all the commands for this module.
func (module ImageModule) Init() {
	RegisterCommand(Command{
		Name:            "image",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
