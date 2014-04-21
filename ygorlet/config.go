// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/jessevdk/go-flags"
)

type Cmd struct {
	ConfigFile string `short:"c" description:"Configuration file" default:"/etc/ygorlet.conf"`
}

type Cfg struct {
	AWSAccessKeyId     string
	AWSSecretAccessKey string

	Name		   string

	// Name of the minion.
	QueueName               string

	// Defines the queue URL for the soul, used to send feedback.
	SoulQueueName      string

	// Region code as found in the AWS API doc (http://goo.gl/Z7KvW), for
	// example: us-east-1.
	AWSRegionCode         string

	// In Test-mode, this program will not attempt to communicate with any
	// external systems (e.g. SQS and will print everything to stdout).
	// Additionally, all delays are reduced to a minimum to speed up the
	// test suite.
	Test bool

	// Points to the filepath of the xxxterm/xombrero socket.
	XombreroSocket string
}

var (
	cfg = Cfg{}
	cmd = Cmd{}
)

// Look in the current directory for an config.json file.
func parseConfigFile() error {
	file, err := os.Open(cmd.ConfigFile)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return err
	}

	if cfg.Name == "" {
		return errors.New("\"Name\" is required")
	}

	if cfg.QueueName == "" {
		return errors.New("\"QueueName\" is required")
	}

	if cfg.SoulQueueName == "" {
		return errors.New("\"SoulQueueName\" is required")
	}

	if cfg.AWSAccessKeyId == "" {
		return errors.New("\"AWSAccessKeyId\" is required")
	}

	if cfg.AWSSecretAccessKey == "" {
		return errors.New("\"AWSSecretAccessKey\" is required")
	}

	if cfg.AWSRegionCode == "" {
		return errors.New("\"AWSRegionCode\" is required")
	}

	return nil
}

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
