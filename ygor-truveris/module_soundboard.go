// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"strconv"
	"time"
)

type SoundBoardModule struct{}

func formatPlayTuneCommand(filename string, duration uint64) string {
	if duration > 0 {
		return fmt.Sprintf("play %s %d", filename, duration)
	} else {
		return fmt.Sprintf("play %s", filename)
	}
}

func (module SoundBoardModule) PrivMsg(msg *PrivMsg) {}

func Play(msg *PrivMsg) {
	var duration uint64 = 0
	var err error

	if len(msg.Args) == 0 {
		privMsg(msg.ReplyTo, "usage: play sound [duration]")
		return
	}

	if len(msg.Args) > 1 {
		duration, err = strconv.ParseUint(msg.Args[1], 10, 8)
		if err != nil {
			duration = 0
		}
	}

	SendToMinion(msg.ReplyTo, formatPlayTuneCommand(msg.Args[0], duration))
}

func PlayAfrica(msg *PrivMsg) {
	msg.Args = append([]string{"tunes/africa.ogg"}, msg.Args...)

	Play(msg)

	go func() {
		time.Sleep(2 * time.Second)
		privAction(msg.ReplyTo, "hears the drums echoing tonight,")
		time.Sleep(5 * time.Second)
		privMsg(msg.ReplyTo, "But she hears only whispers of some quiet conversation")
		time.Sleep(9 * time.Second)
		privMsg(msg.ReplyTo, "She's coming in the 12:30 flight")
		time.Sleep(3 * time.Second)
		privMsg(msg.ReplyTo, "The moonlit wings reflect the stars that guide me towards salvation")
	}()
}

func PlayJeopardy(msg *PrivMsg) {
	msg.Args = append([]string{"tunes/jeopardy.mp3"}, msg.Args...)

	privAction(msg.ReplyTo, "queues some elevator music...")
	Play(msg)
}

func (module SoundBoardModule) Init() {
	RegisterCommand(Command{
		Name:         "play",
		Function:     Play,
		Addressed:    true,
		AllowDirect:  false,
		AllowChannel: true,
	})

	RegisterCommand(Command{
		Name:         "africa",
		Function:     PlayAfrica,
		Addressed:    false,
		AllowDirect:  false,
		AllowChannel: true,
	})

	RegisterCommand(Command{
		Name:         "jeopardy",
		Function:     PlayJeopardy,
		Addressed:    false,
		AllowDirect:  false,
		AllowChannel: true,
	})
}
