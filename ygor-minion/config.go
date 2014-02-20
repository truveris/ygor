// Copyright (c) 2014 Bertrand Janin <b@janin.com>
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

	// In Debug-mode, this program will not attempt to communicate with any
	// external systems (e.g. SQS and will print everything to stdout).
	Debug		   bool

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
