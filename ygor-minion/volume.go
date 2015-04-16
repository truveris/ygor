// Copyright 2014-2015, Truveris Inc. All Rights Reserved.

package main

import (
	"log"
	"os/exec"
	"regexp"
)

var (
	rePercentage = regexp.MustCompile(`^\d+%$`)
	reDecibel    = regexp.MustCompile(`^\d+dB[+-]?$`)
)

// Volume sets the volume via the alsa mixer command.
func Volume(data string) {
	if !rePercentage.MatchString(data) && !reDecibel.MatchString(data) {
		Send("volume error invalid input")
		log.Printf("volume: invalid input '%s'", data)
		return
	}

	cmd := exec.Command(cfg.AMixerCommand, "sset", cfg.AMixerControl, data)
	err := cmd.Start()
	if err != nil {
		Send("volume error starting amixer")
		log.Printf("volume: error starting amixer: %s", err.Error())
		return
	}

	err = cmd.Wait()
	if err != nil {
		Send("volume error during amixer")
		log.Printf("volume: error during amixer: %s", err.Error())
		return
	}

	Send("volume ok")
}
