// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"log"
	"os"
	"time"
)

func main() {
	cmdline := ParseCommandLine()

	cfg, err := ParseConfigFile(cmdline)
	if err != nil {
		log.Fatal("config error: ", err.Error())
	}

	srv := CreateServer(cfg)

	log.Printf("registering modules")
	srv.RegisterModule(&AliasModule{})
	srv.RegisterModule(&CommandsModule{})
	srv.RegisterModule(&ImageModule{})
	srv.RegisterModule(&RebootModule{})
	srv.RegisterModule(&MinionsModule{})
	srv.RegisterModule(&NopModule{})
	srv.RegisterModule(&PingModule{})
	srv.RegisterModule(&PlayModule{})
	srv.RegisterModule(&SayModule{})
	srv.RegisterModule(&SkipModule{})
	srv.RegisterModule(&ShutUpModule{})
	srv.RegisterModule(&TurretModule{})
	srv.RegisterModule(&VolumeModule{})
	srv.RegisterModule(&XombreroModule{})

	log.Printf("starting i/o adapters")
	minionerrch, err := srv.StartAdapters()
	if err != nil {
		log.Fatal("failed to start adapters: ", err.Error())
	}

	go waitForTraceRequest()

	log.Printf("ready")
	for {
		select {
		case err := <-minionerrch:
			log.Printf("minion handler error: %s", err.Error())
		case msg := <-InputQueue:
			switch msg.Type {
			case MsgTypeIRCChannel:
				go srv.IRCMessageHandler(msg)
			case MsgTypeIRCPrivate:
				go srv.IRCMessageHandler(msg)
			case MsgTypeMinion:
				go srv.MinionMessageHandler(msg)
			case MsgTypeExit:
				log.Printf("terminating: %s", msg.Body)
				os.Exit(0)
			case MsgTypeFatal:
				log.Fatal("fatal error: " + msg.Body)
			default:
				log.Printf("msg handler error: un-handled type"+
					" '%d'", msg.Type)
			}

			// This delay allows each scheduled routine to start.
			// This is not particularly pretty but in case a user
			// sends multiple commands in the same request (e.g.
			// say 1; say 2; say 3), it should give enough time for
			// the output messages to be processed in the same
			// order they were received.
			time.Sleep(50 * time.Millisecond)
		}
	}
}
