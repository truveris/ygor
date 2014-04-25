// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// The io_stdio adapter is primarily used for debugging and allows the user to
// interract with ygord from the terminal using stdtin and stdout. The test
// suite uses this adapter extensively since all the tests are based on the
// expectation of an output given a specific input.
//

package main

import (
	"bufio"
	"log"
	"os"
	"strings"
)

// This is used for debugging and local tests.
//
// It fetches queue messages from stdin instead of AWS SQS. It also write to
// stdout anything meant for IRC output.
func StartStdioHandler() (chan error, chan error, error) {
	errch := make(chan error, 0)

	go WriteIRCOutgointToStdout(errch)
	go ReadAllInputFromStdin(errch)

	return errch, nil, nil
}

// Go routine used to write all the IRC output to stdout.
func WriteIRCOutgointToStdout(errch chan error) {
	for {
		line := <-IRCOutgoing
		_, err := os.Stdout.WriteString(line + "\n")
		if err != nil {
			errch <- err
		}
	}
}

// Go routine used to read all ygord input from Stdin (instead of IRC, minions,
// etc.). This is mostly used for debugging/testing.
func ReadAllInputFromStdin(errch chan error) {
	br := bufio.NewReader(os.Stdin)
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			log.Printf("terminating: " + err.Error())
			os.Exit(0)
		}
		line = strings.TrimSpace(line)

		args := strings.SplitN(line, " ", 2)
		if len(args) != 2 {
			log.Printf("not enough elements: %s", args)
			continue
		}

		msgtype := args[0]
		line = args[1]

		switch msgtype {
		case "irc":
			msg := NewMessageFromIRCLine(line)
			if msg != nil {
				InputQueue <- msg
			}
		case "minion":
			args := strings.SplitN(line, " ", 2)
			if len(args) != 2 {
				log.Printf("not enough elements (minion): %s", args)
				continue
			}
			userid := args[0]
			line = args[1]
			msg := NewMessageFromMinionLine(line)
			msg.UserID = userid
			InputQueue <- msg
		default:
			log.Printf("unknown msg type: %s", msgtype)
		}
	}
}
