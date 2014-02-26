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

type Cfg struct {
	AwsAccessKeyId     string
	AwsSecretAccessKey string
	QueueURL           string

	// In Debug-mode, this program will not attempt to communicate with any
	// external systems (e.g. SQS and will print everything to stdout).
	Debug bool

	// Default channel. This is to be replaced by an array of channels
	Channels []string

	// Debug channel. The bot will try to send debug information to this
	// channel in lieu of log file.
	DebugChannel string

	// Who is allowed to do special commands.
	Owner string

	// Any chatter from these nicks will be dropped.
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
		println("error: " + err.Error())
		flagParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}
}

// Look in the current directory for an config.json file.
func parseConfigFile() {
	file, err := os.Open("config.json")
	if err != nil {
		println("error: " + err.Error())
		os.Exit(1)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		println("error: " + err.Error())
		os.Exit(1)
	}
}
