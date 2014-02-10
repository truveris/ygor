// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"time"
	"strings"
)

type SoundBoardModule struct { }

func (module SoundBoardModule) PrivMsg(nick, where, msg string, isAction bool) {
	if strings.HasPrefix(msg, cmd.Nickname+": say ") {
		sendToMinion(msg[6:])
		return
	}

	// Allow both "ygor: tune" and "tune" for those specific keywords.
	msg = strings.Replace(msg, cmd.Nickname+": ", "", 1)

	switch strings.ToLower(msg) {
	case "jeopardy":
		privAction(where, "queues some elevator music...")
		sendToMinion("play-tune jeopardy.mp3")
	case "africa":
		sendToMinion("play-tune africa.ogg")
		time.Sleep(2 * time.Second)
		privAction(where, "hears the drums echoing tonight,")
		time.Sleep(5 * time.Second)
		privMsg(where, "But she hears only whispers of some quiet conversation")
		time.Sleep(9 * time.Second)
		privMsg(where, "She's coming in the 12:30 flight")
		time.Sleep(3 * time.Second)
		privMsg(where, "The moonlit wings reflect the stars that guide me towards salvation")
	case "wagner":
		sendToMinion("play-tune wagner.ogg")
	case "nuke":
		sendToMinion("play-tune nuke_ready.ogg")
	case "energy":
		sendToMinion("play-tune energy.ogg")
	}
}

func (module SoundBoardModule) Init() {
}
