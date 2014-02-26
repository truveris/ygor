// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"github.com/jessevdk/go-flags"
	"os"
)

type Cfg struct {
	Nickname string `long:"nickname" description:"Bot's nickname" default:"ygor"`
	Hostname string `long:"hostname" description:"IRC server to connect to" default:"localhost:6667"`
}

var (
	cfg      = Cfg{}
	soulArgs = []string{}
)

// Parse the command line arguments and return the soul program's path/name
// (only argument).
func parseCommandLine() {
	flagParser := flags.NewParser(&cfg, flags.PassDoubleDash)
	flagParser.Usage = "[OPTIONS] soul-program"

	args, err := flagParser.Parse()
	if err != nil {
		println("error: " + err.Error())
		flagParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	if len(args) < 1 {
		println("error: missing soul program")
		flagParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	soulArgs = args
}
