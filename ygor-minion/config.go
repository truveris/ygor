// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"encoding/json"
	"os"
)

type Cfg struct {
	AwsAccessKeyId     string
	AwsSecretAccessKey string
	QueueURL           string

	// In Test-mode, this program will not attempt to communicate with any
	// external systems (e.g. SQS and will print everything to stdout).
	// Additionally, all delays are reduced to a minimum to speed up the
	// test suite.
	Test bool

	XombreroSocket	   string
}

var (
	cfg = Cfg{}
)

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
