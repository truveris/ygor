// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"log"

	"github.com/truveris/ygor"
)

var (
	Aliases *ygor.AliasFile
)

func main() {
	ParseCommandLine()

	err := ParseConfigFile()
	if err != nil {
		log.Fatal("config error: ", err.Error())
	}

	// We have a global alias file available to everyone. The alias module
	// uses it, the irc io uses it to resolve aliases on PRIVMSGs.
	Aliases, err = ygor.OpenAliasFile(cfg.AliasFilePath)
	if err != nil {
		log.Fatal("alias file error: ", err.Error())
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
