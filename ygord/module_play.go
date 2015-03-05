// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
)

// PlayModule controls the 'play'
type PlayModule struct{}

// PrivMsg is the message handler for 'play' requests.
func (module *PlayModule) PrivMsg(msg *Message) {
	var duration, cmd string

	if len(msg.Args) == 0 {
		IRCPrivMsg(msg.ReplyTo, "usage: play sound [duration]")
		return
	}

	filename := msg.Args[0]
	if len(msg.Args) > 1 {
		duration = msg.Args[1]
	}

	if duration != "" {
		cmd = fmt.Sprintf("play %s %s", filename, duration)
	} else {
		cmd = fmt.Sprintf("play %s", filename)
	}

	SendToChannelMinions(msg.ReplyTo, cmd)
}

// Init registers all the commands for this module.
func (module *PlayModule) Init() {
	RegisterCommand(Command{
		Name:            "play",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
