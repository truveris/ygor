// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"time"

	"github.com/truveris/ygor"
)

type SoundBoardModule struct{}

func formatPlayTuneCommand(filename, duration string) string {
	if duration != "" {
		return fmt.Sprintf("play %s %s", filename, duration)
	} else {
		return fmt.Sprintf("play %s", filename)
	}
}

func (module SoundBoardModule) PrivMsg(msg *ygor.Message) {}

func Play(msg *ygor.Message) {
	var duration string

	if len(msg.Args) == 0 {
		IRCPrivMsg(msg.ReplyTo, "usage: play sound [duration]")
		return
	}

	filename := msg.Args[0]
	if len(msg.Args) > 1 {
		duration = msg.Args[1]
	}

	SendToChannelMinions(msg.ReplyTo, formatPlayTuneCommand(filename, duration))
}

func PlayAfrica(msg *ygor.Message) {
	msg.Args = append([]string{"tunes/africa.ogg"}, msg.Args...)

	Play(msg)

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

func (module SoundBoardModule) Init() {
	ygor.RegisterCommand(ygor.Command{
		Name:            "play",
		PrivMsgFunction: Play,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	ygor.RegisterCommand(ygor.Command{
		Name:            "africa",
		PrivMsgFunction: PlayAfrica,
		Addressed:       false,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
