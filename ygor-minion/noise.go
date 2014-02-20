// Copyright 2014, Truveris Inc. All Rights Reserved.

package main

import (
	"log"
	"os"
)

var (
	RunningProcess *os.Process
)

type Noise struct {
	Path     string
	Duration string
	Sentence string
}

func playTune(tune Noise) {
	cmd := mplayer(tune)
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
