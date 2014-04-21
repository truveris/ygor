// Copyright 2014, Truveris Inc. All Rights Reserved.

package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/tamentis/go-mplayer"
)

var (
	RunningProcess *os.Process

	// This is the noise box, we keep as much as possible in local memory,
	// that makes 'shutup' remotely useful. Without buffer we would have to
	// wait through orders before even reaching the 'shutup' command.
	NoiseInbox = make(chan Noise, 1000)
)

type Noise struct {
	Path     string
	Duration time.Duration
	Sentence string
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

func playNoise(noiseInbox chan Noise) {
	for noise := range noiseInbox {
		if noise.Sentence != "" {
			say(noise.Sentence)
		} else {
			playTune(noise)
		}
	}
}

// Empty the noise inbox ...
func ShutUp() {
	log.Printf("shutup: deleting %d items from the noise queue",
		len(NoiseInbox))

	if len(NoiseInbox) > 0 {
		for _ = range NoiseInbox {
			if len(NoiseInbox) == 0 {
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

func Say(data string) {
	noise := Noise{}
	noise.Sentence = data
	NoiseInbox <- noise
}

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
			SendToSoul("play error invalid duration: " + err.Error())
			return
		}
		tune.Duration = duration
	}
	NoiseInbox <- tune
}
