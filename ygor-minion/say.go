// Copyright 2014-2015, Truveris Inc. All Rights Reserved.

package main

import (
	"log"
	"net/url"
	"os/exec"
	"runtime"
	"strings"

	"github.com/jessevdk/go-flags"
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

	cmdEspeak := exec.Command("espeak", "-ven-us+f2", "--stdout",
		sentence, "-a", "300", "-s", "130")
	cmdAplay := exec.Command("aplay")

	cmdAplay.Stdin, err = cmdEspeak.StdoutPipe()
	if err != nil {
		log.Printf("error on cmdEspeak.StdoutPipe(): " + err.Error())
		return
	}

	err = cmdEspeak.Start()
	if err != nil {
		log.Printf("error on cmdEspeak.Start(): " + err.Error())
		return
	}
	err = cmdAplay.Start()
	if err != nil {
		log.Printf("error on cmdAplay.Start(): " + err.Error())
		return
	}

	RunningProcess = cmdAplay.Process

	err = cmdEspeak.Wait()
	if err != nil {
		log.Printf("error on cmdEspeak.Wait(): " + err.Error())
		return
	}
	err = cmdAplay.Wait()
	if err != nil {
		log.Printf("error on cmdAplay.Wait(): " + err.Error())
		return
	}

	RunningProcess = nil
}

// Say is the implementation of the 'say' minion command.
func Say(data string) {
	cmd := SayArgs{}
	args := strings.Split(data, " ")

	flagParser := flags.NewParser(&cmd, flags.PassDoubleDash)
	args, err := flagParser.ParseArgs(args)
	if err != nil {
		log.Printf("say command line error: %s", err.Error())
		return
	}

	noise := Noise{}
	noise.Voice = cmd.Voice
	noise.Sentence = strings.Join(args, " ")
	AddToPlayList(noise)
}
