// Copyright 2014, Truveris Inc. All Rights Reserved.
//
// ygor-minion takes orders from ygor and executes them (through an SQS
// queue/inbox). There could be hundreds of minions installed on different
// machines, they can all have different purposes, it's up to ygor to decide.
//
// Messages to ygor-minion should be short and sweet, with nothing but plain
// text. They should take the form of a command and its parameters, for
// example:
//
// 	play valkyries.mp3
//
// The cost of one ygor-minion in SQS is less than a dollar a year, at one
// query per 20 seconds:
//
// 	Number of requests per day: (60 * 60 * 24) / 20 = 4320
// 	Number of requests per year: 4320 * 365 = 1576800
// 	Cost per request: $0.0000005
// 	Total cost per year: 1576800 * 0.0000005 = $0.7884
//

package main

import (
	"bufio"
	"fmt"
	"runtime"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
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

// say (for macs)
func macSay(sentence string) {
	cmd := exec.Command("say", sentence)
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

// espeak | aplay (for linux)
func say(sentence string) {
	var err error

	if cfg.Debug {
		log.Printf("say(%s)", sentence)
		return
	}

	if runtime.GOOS == "darwin" {
		macSay(sentence)
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

func playNoise(noiseInbox chan Noise) {
	for noise := range noiseInbox {
		if noise.Sentence != "" {
			say(noise.Sentence)
		} else {
			playTune(noise)
		}
	}
}

// This is used for debugging.
//
// It fetches queue messages from stdin instead of AWS SQS.
//
func fetchMessagesFromStdin(incoming chan string) {
	br := bufio.NewReader(os.Stdin)

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			log.Fatal("terminating: " + err.Error())
		}
		line = strings.TrimSpace(line)

		incoming <- line
	}
}

func fetchMessages(incoming chan string) {
	for {
		body, receipt, err := getMessage()
		if err != nil {
			log.Printf("error: %s", err.Error())
			time.Sleep(10 * time.Second)
		}

		if body == "" {
			continue
		}

		deleteMessage(receipt)

		incoming <- body
	}
}

// Loop until the end of time.
//
// In case of error, delay the next loop. Automatically reconnect if everything
// goes fine (for 0 or 1 message).
func main() {
	if len(os.Args) != 1 {
		fmt.Printf("usage: ygor-minion\n")
		os.Exit(1)
	}

	parseConfigFile()

	// This is the message box.
	incoming := make(chan string)
	if cfg.Debug {
		go fetchMessagesFromStdin(incoming)
	} else {
		go fetchMessages(incoming)
	}

	// This is the noise box, we keep as much as possible in local memory,
	// that makes 'shutup' remotely useful. Without buffer we would have to
	// wait through orders before even reaching the 'shutup' command.
	noiseInbox := make(chan Noise, 1000)
	go playNoise(noiseInbox)

	log.Printf("ygor-minion ready!")

	for body := range incoming {
		log.Printf("got message: \"%s\"", body)

		tokens := strings.Split(body, " ")
		switch tokens[0] {
		case "play", "play-tune":
			if len(tokens) > 1 {
				tune := Noise{}
				tune.Path = tokens[1]
				if len(tokens) > 2 {
					tune.Duration = tokens[2]
				}
				noiseInbox <- tune
			}
		case "say":
			noise := Noise{}
			noise.Sentence = strings.Join(tokens[1:], " ")
			noiseInbox <- noise
		case "shutup":
			// Empty the noise inbox ...
			log.Printf("deleting %d items from the noise queue",
				len(noiseInbox))
			if len(noiseInbox) > 0 {
				for _ = range noiseInbox {
					if len(noiseInbox) == 0 {
						break
					}
				}
			}

			// ... then kill whatever is currently running.
			if RunningProcess != nil {
				if err := RunningProcess.Kill(); err != nil {
					log.Printf("error trying to kill "+
						"current process: %s",
						err.Error())
				}
			}
		default:
			log.Printf("unknown command %s", tokens[0])
		}
	}
}
