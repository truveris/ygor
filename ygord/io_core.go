// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// The io_* portion of the ygord code base defines all the adapters feeding
// data to ygor (e.g. irc-sqs, minion-sqs, stdin).
//

package main

import (
	"errors"

	"github.com/truveris/sqs"
	"github.com/truveris/ygor"
)

var (
	// All the normalized messages are pushed by the IO adapters, only the
	// main loop reads from there.
	InputQueue = make(chan *ygor.Message)
)

// Start all the IO adapters (IRC, Stdin/Stdout, Minions, API, etc.)
func StartAdapters() (<-chan error, <-chan error, error) {
	err := StartHTTPAdapter()
	if err != nil {
		return nil, nil, errors.New("error starting http adapter: " +
			err.Error())
	}

	if cfg.TestMode {
		return StartStdioHandler()
	}

	client, err := sqs.NewClient(cfg.AWSAccessKeyId, cfg.AWSSecretAccessKey,
		cfg.AWSRegionCode)
	if err != nil {
		return nil, nil, err
	}

	ircerrch, err := StartIRCAdapter(client)
	if err != nil {
		return nil, nil, errors.New("error starting IRC adapter: " +
			err.Error())
	}

	minionerrch, err := StartMinionAdapter(client)
	if err != nil {
		return nil, nil, errors.New("error starting minion adapter: " +
			err.Error())
	}

	return ircerrch, minionerrch, nil
}
