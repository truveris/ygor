// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"log"
	"os"

	"github.com/truveris/ygor"
)

var (
	Aliases *ygor.AliasFile
	Minions *ygor.MinionsFile
)

func main() {
	ParseCommandLine()

	err := ParseConfigFile()
	if err != nil {
		log.Fatal("config error: ", err.Error())
	}

	// We have global alias and minions files available to everyone. The
	// alias module and irc io adapter use aliases and everything uses minions.
	Aliases, err = ygor.OpenAliasFile(cfg.AliasFilePath)
	if err != nil {
		log.Fatal("alias file error: ", err.Error())
	}

	Minions, err = ygor.OpenMinionsFile(cfg.MinionsFilePath)
	if err != nil {
		log.Fatal("minions file error: ", err.Error())
	}

	log.Printf("registering modules")
	RegisterModule(&AliasModule{})
	RegisterModule(&ImageModule{})
	RegisterModule(&RebootModule{})
	RegisterModule(&MinionsModule{})
	RegisterModule(&NopModule{})
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

	go ygor.WaitForTraceRequest()

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
				go IRCMessageHandler(msg)
			case ygor.MsgTypeIRCPrivate:
				go IRCMessageHandler(msg)
			case ygor.MsgTypeMinion:
				go MinionMessageHandler(msg)
			case ygor.MsgTypeExit:
				log.Printf("terminating: %s", msg.Body)
				os.Exit(0)
			case ygor.MsgTypeFatal:
				log.Fatal("fatal error: " + msg.Body)
			default:
				log.Printf("msg handler error: un-handled type"+
					" '%d'", msg.Type)
			}
		}
	}
}
