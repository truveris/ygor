// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"log"
	"os"
	"time"

	"github.com/truveris/ygor/ygord/alias"
)

// These are the global registry for aliases and minions.  Both are in-memory
// maps with permanent file storage.
var (
	Aliases *alias.File
	Minions *MinionsFile
)

func main() {
	ParseCommandLine()

	err := ParseConfigFile()
	if err != nil {
		log.Fatal("config error: ", err.Error())
	}

	// We have global alias and minions files available to everyone. The
	// alias module and irc io adapter use aliases and everything uses minions.
	Aliases, err = alias.Open(cfg.AliasFilePath)
	if err != nil {
		log.Fatal("alias file error: ", err.Error())
	}

	Minions, err = OpenMinionsFile(cfg.MinionsFilePath)
	if err != nil {
		log.Fatal("minions file error: ", err.Error())
	}

	log.Printf("registering modules")
	RegisterModule(&AliasModule{})
	RegisterModule(&CommandsModule{})
	RegisterModule(&ImageModule{})
	RegisterModule(&RebootModule{})
	RegisterModule(&MinionsModule{})
	RegisterModule(&NopModule{})
	RegisterModule(&PingModule{})
	RegisterModule(&PlayModule{})
	RegisterModule(&SayModule{})
	RegisterModule(&SkipModule{})
	RegisterModule(&ShutUpModule{})
	RegisterModule(&TurretModule{})
	RegisterModule(&VolumeModule{})
	RegisterModule(&XombreroModule{})

	log.Printf("starting i/o adapters")
	ircerrch, minionerrch, err := StartAdapters()
	if err != nil {
		log.Fatal("failed to start adapters: ", err.Error())
	}

	go waitForTraceRequest()

	log.Printf("ready")
	for {
		select {
		case err := <-ircerrch:
			log.Printf("irc handler error: %s", err.Error())
		case err := <-minionerrch:
			log.Printf("minion handler error: %s", err.Error())
		case msg := <-InputQueue:
			switch msg.Type {
			case MsgTypeIRCChannel:
				go IRCMessageHandler(msg)
			case MsgTypeIRCPrivate:
				go IRCMessageHandler(msg)
			case MsgTypeMinion:
				go MinionMessageHandler(msg)
			case MsgTypeExit:
				log.Printf("terminating: %s", msg.Body)
				os.Exit(0)
			case MsgTypeFatal:
				log.Fatal("fatal error: " + msg.Body)
			default:
				log.Printf("msg handler error: un-handled type"+
					" '%d'", msg.Type)
			}
			time.Sleep(50 * time.Millisecond)
		}
	}
}
