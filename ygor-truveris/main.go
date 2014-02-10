// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	channel = "#dev"
	owner   = "b"

	// Used to synchronize line output.
	outgoing = make(chan string)

	// Indicate an EOF on stdin.
	eof = make(chan string)

	// All the modules currently registered.
	modules = make([]Module, 0)

	// Detect input patterns.
	rePrivMsg = regexp.MustCompile(`^:([^!]+)![^@]+@[^\s]+\sPRIVMSG\s([^\s]+)\s:(.*)`)
)

type Module interface {
	PrivMsg(nick, where, msg string, isAction bool)
	Init()
}

// Print whatever arrives on the channel to stdout. It allows all the
// modules to send lines to the IRC server without the risk of mangling
// them.
//
// Modules should be pushing lines without new-line characters.
//
// This functions adds 10 ms per character to make the output speed a little
// more natural.
func outgoingHandler() {
	for msg := range outgoing {
		time.Sleep(time.Duration(len(msg)*10) * time.Millisecond)
		io.WriteString(os.Stdout, msg)
		io.WriteString(os.Stdout, "\n")
	}
}

// Feed the incoming channel with lines read from stdin.
func incomingHandler() {
	br := bufio.NewReader(os.Stdin)
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			eof <- err.Error()
		}
		line = strings.TrimSpace(line)

		// PRIVMSG
		tokens := rePrivMsg.FindStringSubmatch(line)
		if tokens != nil {
			nick, where, msg := tokens[1], tokens[2], tokens[3]
			isAction := false
			if strings.HasPrefix(msg, "\x01ACTION ") {
				msg = msg[8 : len(msg)-1]
				isAction = true
			}
			for _, module := range modules {
				module.PrivMsg(nick, where, msg, isAction)
			}
			continue
		}
	}
}

// Send a message to a channel.
func privMsg(channel, msg string) {
	lines := strings.Split(msg, "\n")
	for i := 0; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}
		outgoing <- fmt.Sprintf("PRIVMSG %s :%s", channel, lines[i])
		time.Sleep(500 * time.Millisecond)
	}
}

// Send an action message to a channel.
func privAction(channel, msg string) {
	outgoing <- fmt.Sprintf("PRIVMSG %s :\x01ACTION %s\x01", channel, msg)
}

func registerModule(module Module) {
	module.Init()
	modules = append(modules, module)
}

func main() {
	parseCommandLine()
	parseConfigFile()

	//	if cfg.MarkovDataPath != "" {
	//		chain := initializeMarkovChain()
	//		registerModule(RandomModule{Markov: chain})
	//	}
	//
	//	registerModule(RepeatModule{})
	//	registerModule(NeeextModule{})

	registerModule(SoundBoardModule{})

	go outgoingHandler()
	go incomingHandler()

	outgoing <- "JOIN " + channel

	panic(<-eof)
}
