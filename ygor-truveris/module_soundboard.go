// Copyright (c) 2014 Bertrand Janin <b@janin.com>
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

func Play(where string, params []string) {
	var duration uint64 = 0
	var err error

	if len(params) == 0 {
		privMsg(where, "usage: play sound [duration]")
		return
	}

	if len(params) > 1 {
		duration, err = strconv.ParseUint(params[1], 10, 8)
		if err != nil {
			duration = 0
		}
	}

	SendToMinion(formatPlayTuneCommand(params[0], duration))
}

func PlayAfrica(where string, params []string) {
	params = append([]string{"tunes/africa.ogg"}, params...)

	Play(where, params)

	go func() {
		time.Sleep(2 * time.Second)
		privAction(where, "hears the drums echoing tonight,")
		time.Sleep(5 * time.Second)
		privMsg(where, "But she hears only whispers of some quiet conversation")
		time.Sleep(9 * time.Second)
		privMsg(where, "She's coming in the 12:30 flight")
		time.Sleep(3 * time.Second)
		privMsg(where, "The moonlit wings reflect the stars that guide me towards salvation")
	}()
}

func PlayJeopardy(where string, params []string) {
	params = append([]string{"tunes/jeopardy.mp3"}, params...)

	privAction(where, "queues some elevator music...")
	Play(where, params)
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
