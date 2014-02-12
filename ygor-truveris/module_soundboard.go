// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"regexp"
	"strings"
	"time"
	"strconv"
	"fmt"
)

var (
	reSoundBoard = regexp.MustCompile(`^(?:\w+ )?(\w+)(?:(?:\s+for)?\s+([0-9]+))?`)
	reStop = regexp.MustCompile(`^st[aho]+p`)
	reShhh = regexp.MustCompile(`^s+[sh]+`)
)

type SoundBoardModule struct{}

func getTune(msg string) (string, uint64) {
	tokens := reSoundBoard.FindStringSubmatch(msg)
	if tokens == nil {
		return "", 0
	}

	tune := strings.ToLower(tokens[1])
	duration, err := strconv.ParseUint(tokens[2], 10, 8)
	if err != nil {
		// That really shouldn't happen since the regexp should only
		// capture uint but we're being cautious.
		return tune, 0
	}

	return tune, duration
}

func formatPlayTuneCommand(filename string, duration uint64) string {
	if duration > 0 {
		return fmt.Sprintf("play-tune %s %d", filename, duration)
	} else {
		return fmt.Sprintf("play-tune %s", filename)
	}
}

func playTune(where string, tune string, duration uint64) {
	switch tune {
	case "jeopardy":
		privAction(where, "queues some elevator music...")
		sendToMinion(formatPlayTuneCommand("jeopardy.mp3", duration))
	case "africa":
		sendToMinion(formatPlayTuneCommand("africa.ogg", duration))
		time.Sleep(2 * time.Second)
		privAction(where, "hears the drums echoing tonight,")
		time.Sleep(5 * time.Second)
		privMsg(where, "But she hears only whispers of some quiet conversation")
		time.Sleep(9 * time.Second)
		privMsg(where, "She's coming in the 12:30 flight")
		time.Sleep(3 * time.Second)
		privMsg(where, "The moonlit wings reflect the stars that guide me towards salvation")
	case "wagner":
		sendToMinion(formatPlayTuneCommand("wagner.ogg", duration))
	case "nuke":
		sendToMinion(formatPlayTuneCommand("nuke_ready.ogg", duration))
	case "energy":
		sendToMinion(formatPlayTuneCommand("energy.ogg", duration))
	}
}

func isShutUpRequest(msg string) bool {
	msg = strings.ToLower(msg)
	if reStop.MatchString(msg) || reShhh.MatchString(msg) {
		return true
	}
	return false
}

func shutup(where string) {
	sendToMinion("shutup")
	privMsg(where, "ok...")
}

func (module SoundBoardModule) PrivMsg(nick, where, msg string, isAction bool) {
	if msg = AddressedToMe(msg); msg == "" {
		return
	}

	if isShutUpRequest(msg) {
		shutup(where)
		return
	}

	tune, duration := getTune(msg)
	if tune == "" {
		return
	}

	playTune(where, tune, duration)
}

func (module SoundBoardModule) Init() {
}
