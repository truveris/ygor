// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// The io_* portion of the ygord code base defines all the adapters feeding
// data to ygor (e.g. irc-sqs, minion-sqs, stdin).
//

package main

import (
	"errors"

	"github.com/truveris/sqs"
)

var (
	// InputQueue is the channel used to receive all the normalized
	// messages pushed by the IO adapters, only the main loop reads it.
	InputQueue = make(chan *Message)
)

// StartAdapters starts all the IO adapters (IRC, Stdin/Stdout, Minions, API)
func StartAdapters() (<-chan error, error) {
	err := StartHTTPAdapter()
	if err != nil {
		return nil, errors.New("error starting http adapter: " +
			err.Error())
	}

	client, err := sqs.NewClient(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey,
		cfg.AWSRegionCode)
	if err != nil {
		return nil, err
	}

	err = StartIRCAdapter()
	if err != nil {
		return nil, errors.New("error starting IRC adapter: " +
			err.Error())
	}

	minionerrch, err := StartMinionAdapter(client)
	if err != nil {
		return nil, errors.New("error starting minion adapter: " +
			err.Error())
	}

	return minionerrch, nil
}
