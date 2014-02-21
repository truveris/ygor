// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

var (
	// Used to synchronize line output.
	outgoing = make(chan string)

	// Indicate an EOF on stdin.
	eof = make(chan string)

	// All the modules currently registered.
	modules = make([]Module, 0)
)

type Module interface {
	PrivMsg(msg *PrivMsg)
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
		io.WriteString(os.Stdout, msg)
		io.WriteString(os.Stdout, "\n")
	}
}

// Feed the incoming channel with lines read from stdin.
func incomingHandler() {
	br := bufio.NewReader(os.Stdin)

	OuterLoop:
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			eof <- err.Error()
		}
		line = strings.TrimSpace(line)

		msg := NewPrivMsg(line)
		if msg == nil {
			continue
		}

		// Check if we should ignore this message.
		for _, ignore := range cfg.Ignore {
			if ignore == msg.Nick {
				continue OuterLoop
			}
		}

		for _, cmd := range RegisteredCommands {
			if cmd.IsAddressed != msg.IsAddressed {
				continue
			}
			if cmd.Name != msg.Command {
				continue
			}
			cmd.Function(msg.Channel, msg.Args)
			continue OuterLoop
		}

		for _, module := range modules {
			module.PrivMsg(msg)
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

func delayedPrivMsg(channel, msg string, waitTime time.Duration) {
	time.Sleep(waitTime)
	privMsg(channel, msg)
}

func delayedPrivAction(channel, msg string, waitTime time.Duration) {
	time.Sleep(waitTime)
	privAction(channel, msg)
}

func registerModule(module Module) {
	module.Init()
	modules = append(modules, module)
}

func main() {
	parseCommandLine()
	parseConfigFile()

	// if cfg.MarkovDataPath != "" {
	// 	chain := initializeMarkovChain()
	// 	registerModule(RandomModule{Markov: chain})
	// }
	//
	// registerModule(RepeatModule{})
	// registerModule(NeeextModule{})

	registerModule(AliasModule{})
	registerModule(SayModule{})
	registerModule(ShutUpModule{})
	registerModule(SoundBoardModule{})
	registerModule(XombreroModule{})

	go outgoingHandler()
	go incomingHandler()

	// Auto-join all the configured channels.
	for _, channel := range cfg.Channels {
		outgoing <- "JOIN " + channel
	}

	fmt.Fprintf(os.Stderr, "terminating: %s\n", <-eof)
}
