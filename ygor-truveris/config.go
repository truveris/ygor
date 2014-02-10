// Copyright (c) 2014 Bertrand Janin <b@janin.com>
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
