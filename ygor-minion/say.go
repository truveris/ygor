// Copyright 2014, Truveris Inc. All Rights Reserved.

package main

import (
	"log"
	"net/url"
	"os/exec"
	"runtime"
)

// say (for macs)
func macSay(voice, sentence string) {
	cmd := exec.Command("say", "-v", voice, sentence)
	err := cmd.Start()
	if err != nil {
		log.Printf("error starting say")
	}

	RunningProcess = cmd.Process

	err = cmd.Wait()
	if err != nil {
		log.Printf("error on cmd.Wait(): " + err.Error())
		return
	}
}

func sayd(voice, sentence string) {
	url := cfg.SaydURL + voice + "?" + url.QueryEscape(sentence)
	log.Printf("invoking remote sayd with: %s", url)

	if cfg.TestMode {
		return
	}

	mplayerPlayAndWait(url)
}

// Call sayd if configured, if not, "espeak" on Linux and "say" on Mac. The
// voice argument is only used by "sayd" and "say".
func say(voice, sentence string) {
	var err error

	if cfg.SaydURL != "" {
		sayd(voice, sentence)
		return
	}

	if cfg.TestMode {
		log.Printf("say: %s", sentence)
		return
	}

	if runtime.GOOS == "darwin" {
		macSay(voice, sentence)
		return
	}

	cmd_espeak := exec.Command("espeak", "-ven-us+f2", "--stdout",
		sentence, "-a", "300", "-s", "130")
	cmd_aplay := exec.Command("aplay")

	cmd_aplay.Stdin, err = cmd_espeak.StdoutPipe()
	if err != nil {
		log.Printf("error on cmd_espeak.StdoutPipe(): " + err.Error())
		return
	}

	err = cmd_espeak.Start()
	if err != nil {
		log.Printf("error on cmd_espeak.Start(): " + err.Error())
		return
	}
	err = cmd_aplay.Start()
	if err != nil {
		log.Printf("error on cmd_aplay.Start(): " + err.Error())
		return
	}

	RunningProcess = cmd_aplay.Process

	err = cmd_espeak.Wait()
	if err != nil {
		log.Printf("error on cmd_espeak.Wait(): " + err.Error())
		return
	}
	err = cmd_aplay.Wait()
	if err != nil {
		log.Printf("error on cmd_aplay.Wait(): " + err.Error())
		return
	}

	RunningProcess = nil
}
