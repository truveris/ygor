// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"time"
)

// PlayModule controls the 'play' and 'africa' commands.
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

// PrivMsgAfrica is the message handler for 'africa' requests.
func (module *PlayModule) PrivMsgAfrica(msg *Message) {
	msg.Args = append([]string{"tunes/africa.ogg"}, msg.Args...)

	module.PrivMsg(msg)

	go func() {
		time.Sleep(2 * time.Second)
		IRCPrivAction(msg.ReplyTo, "hears the drums echoing tonight,")
		time.Sleep(5 * time.Second)
		IRCPrivMsg(msg.ReplyTo, "But she hears only whispers of some quiet conversation")
		time.Sleep(9 * time.Second)
		IRCPrivMsg(msg.ReplyTo, "She's coming in the 12:30 flight")
		time.Sleep(3 * time.Second)
		IRCPrivMsg(msg.ReplyTo, "The moonlit wings reflect the stars that guide me towards salvation")
	}()
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

	RegisterCommand(Command{
		Name:            "africa",
		PrivMsgFunction: module.PrivMsgAfrica,
		Addressed:       false,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
