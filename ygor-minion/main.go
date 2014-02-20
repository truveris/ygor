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
	"log"
	"net"
	"os"
	"strings"
	"time"
)

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

// Separate command and data.
func SplitBody(body string) (string, string) {
	var command, data string

	tokens := strings.SplitN(body, " ", 2)

	command = tokens[0]
	if len(tokens) > 1 {
		data = tokens[1]
	}

	return command, data
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

		command, data := SplitBody(body)

		switch command {
		case "play", "play-tune":
			if data != "" {
				tokens := strings.Split(data, " ")
				tune := Noise{}
				tune.Path = tokens[0]
				if len(tokens) > 1 {
					tune.Duration = tokens[1]
				}
				noiseInbox <- tune
			}
		case "say":
			noise := Noise{}
			noise.Sentence = data
			noiseInbox <- noise
		case "xombrero":
			// Send a message to xombrero via its unix socket.
			conn, err := net.Dial("unix", cfg.XombreroSocket)
			if err != nil {
				log.Printf("unable to connect to xombrero: %s",
					err.Error())
				continue
			}
			fmt.Fprintf(conn, "%s", data)
			conn.Close()
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
			log.Printf("unknown command %s", command)
		}
	}
}
