// Copyright 2014-2015, Truveris Inc. All Rights Reserved.

package main

import (
	"log"
	"strings"
	"time"

	"github.com/tamentis/go-mplayer"
)

var (
	// This is the noise box, we keep as much as possible in local memory,
	// that makes 'shutup' remotely useful. Without buffer we would have to
	// wait through orders before even reaching the 'shutup' command.
	playlist = make(chan Noise, 64)
)

// Noise defines a sound emitted by our minion. It could be a sound clip, a
// video or a spoken sentence.
type Noise struct {
	Path     string
	Duration time.Duration
	Voice    string
	Sentence string
}

// SayArgs is the definition of the command-line parameters for the say command.
type SayArgs struct {
	Voice string `short:"v" description:"Voice for say" default:"bruce"`
}

func playTune(tune Noise) {
	cmd := player(tune)
	if cmd == nil {
		return
	}

	err := cmd.Start()
	if err != nil {
		log.Printf("error on mplayer Start(): %s", err.Error())
	}

	RunningProcess = cmd.Process

	err = cmd.Wait()
	if err != nil {
		log.Printf("error on mplayer Wait(): %s", err.Error())
	}

	RunningProcess = nil
}

// Iterate over the noise channel and pass the content to "say" or "play".
func playNoise() {
	for noise := range playlist {
		if noise.Sentence != "" {
			say(noise.Voice, noise.Sentence)
		} else {
			playTune(noise)
		}
	}
}

// ShutUp clears the noise inbox.
func ShutUp() {
	log.Printf("shutup: deleting %d items from the noise queue",
		len(playlist))

	if len(playlist) > 0 {
		for _ = range playlist {
			if len(playlist) == 0 {
				break
			}
		}
	}

	mplayer.SendCommand("stop")

	// ... then kill whatever is currently running.
	if RunningProcess != nil {
		if err := RunningProcess.Kill(); err != nil {
			log.Printf("error trying to kill "+
				"current process: %s",
				err.Error())
		}
	}
}

// Skip to the next entry.
func Skip() {
	log.Printf("skip")
	mplayer.Skip()
}

// AddToPlayList pushes noise down the pipeline.
func AddToPlayList(n Noise) {
	playlist <- n
}

// Play is the implementation of the 'play' minion command which schedules the
// playback of sounds and videos in the minion.
func Play(data string) {
	if data == "" {
		return
	}

	tokens := strings.Split(data, " ")
	tune := Noise{}
	tune.Path = tokens[0]
	if len(tokens) > 1 {
		duration, err := time.ParseDuration(tokens[1])
		if err != nil {
			Send("play warning invalid duration: " + err.Error())
		} else {
			tune.Duration = duration
		}
	}
	AddToPlayList(tune)
}
