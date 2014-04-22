// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"errors"
	"log"

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

func main() {
	ParseCommandLine()

	err := ParseConfigFile()
	if err != nil {
		log.Fatal("config error: ", err.Error())
	}

	log.Printf("registering modules")

	RegisterModule(&AliasModule{})
	RegisterModule(&ImageModule{})
	RegisterModule(&RebootModule{})
	RegisterModule(&MinionsModule{})
	RegisterModule(&PingModule{})
	RegisterModule(&SayModule{})
	RegisterModule(&ShutUpModule{})
	RegisterModule(&SoundBoardModule{})
	RegisterModule(&XombreroModule{})

	log.Printf("starting i/o adapters")
	ircerrch, minionerrch, err := StartAdapters()
	if err != nil {
		log.Fatal("failed to start adapters: ", err.Error())
	}

	log.Printf("ready")
	for {
		select {
		case err := <-ircerrch:
			log.Printf("irc handler error: %s", err.Error())
		case err := <-minionerrch:
			log.Printf("minion handler error: %s", err.Error())
		case msg := <-InputQueue:
			switch msg.Type {
			case ygor.MsgTypeIRCChannel:
				IRCMessageHandler(msg)
			case ygor.MsgTypeIRCPrivate:
				IRCMessageHandler(msg)
			case ygor.MsgTypeMinion:
				MinionMessageHandler(msg)
			default:
				log.Printf("msg handler error: un-handled type"+
					" '%d'", msg.Type)
			}
		}
	}
}
