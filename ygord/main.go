// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"log"
	"os"
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
	srv.RegisterModule(&NopModule{})
	srv.RegisterModule(&PingModule{})
	srv.RegisterModule(&PlayModule{})
	srv.RegisterModule(&SayModule{})
	srv.RegisterModule(&SkipModule{})
	srv.RegisterModule(&ShutUpModule{})
	srv.RegisterModule(&VolumeModule{})

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
		case msg := <-srv.InputQueue:
			switch msg.Type {
			case MsgTypeIRCChannel:
				srv.IRCMessageHandler(msg)
			case MsgTypeIRCPrivate:
				srv.IRCMessageHandler(msg)
			case MsgTypeExit:
				log.Printf("terminating: %s", msg.Body)
				os.Exit(0)
			case MsgTypeFatal:
				log.Fatal("fatal error: " + msg.Body)
			default:
				log.Printf("msg handler error: un-handled type"+
					" '%d'", msg.Type)
			}
		case msg := <-srv.OutputQueue:
			switch msg.Type {
			case OutMsgTypePrivMsg:
				conn.Privmsg(msg.Channel, msg.Body)
			case OutMsgTypeAction:
				conn.Action(msg.Channel, msg.Body)
			default:
				log.Printf("outmsg handler error: un-handled type"+
					" '%d'", msg.Type)
			}
		}
	}
}
