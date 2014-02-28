// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"encoding/json"
	"github.com/jessevdk/go-flags"
	"os"
)

type Cmd struct {
	Nickname string `long:"nickname" description:"Bot's nickname" default:"ygor"`
}

type ChannelCfg struct {
	QueueURL           string
}

type Cfg struct {
	// Credentials used to write to the minion queues and read from the
	// soul queue.
	AwsAccessKeyId     string
	AwsSecretAccessKey string

	// In Debug-mode, this program will not attempt to communicate with any
	// external systems (e.g. SQS and will print everything to stdout).
	// Additionally, all delays are reduced to a minimum to speed up the
	// test suite.
	Test bool

	// All the configured channels. The soul will JOIN every single one of
	// them and will push minion commands to the configured associated
	// minion.
	Channels map[string]ChannelCfg

	// Try to send debug information to this channel in lieu of log file.
	DebugChannel string

	// Who is allowed to do special commands.
	Owner string

	// Any chatter from these nicks will be dropped (other bots).
	Ignore []string
}

var (
	cmd = Cmd{}
	cfg = Cfg{}
)

// Parse the command line arguments and return the soul program's path/name
// (only argument).
func parseCommandLine() {
	flagParser := flags.NewParser(&cmd, flags.PassDoubleDash)
	_, err := flagParser.Parse()
	if err != nil {
		println("command line error: " + err.Error())
		flagParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}
}

// Look in the current directory for an config.json file.
func parseConfigFile() {
	file, err := os.Open("config.json")
	if err != nil {
		println("config error: " + err.Error())
		os.Exit(1)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		println("config error: " + err.Error())
		os.Exit(1)
	}
}
