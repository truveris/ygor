// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"log"
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

	err = srv.StartHTTPServer(cfg.HTTPServerAddress)
	if err != nil {
		log.Fatal("failed to start http server: ", err.Error())
	}

	err = srv.StartIRCClient()
	if err != nil {
		log.Fatal("failed to start IRC client: ", err.Error())
	}

	go waitForTraceRequest()

	log.Printf("ready, entering main loop")
	for {
		select {
		case msg := <-srv.IRCInputQueue:
			switch msg.Type {
			case IRCInputMsgTypeIRCChannel:
				srv.IRCMessageHandler(msg)
			case IRCInputMsgTypeIRCPrivate:
				srv.IRCMessageHandler(msg)
			default:
				log.Printf("main loop: un-handled "+
					"IRC input message type"+
					" '%d'", msg.Type)
			}
		case msg := <-srv.IRCOutputQueue:
			switch msg.Type {
			case IRCOutMsgTypePrivMsg:
				conn.Privmsg(msg.Channel, msg.Body)
			case IRCOutMsgTypeAction:
				conn.Action(msg.Channel, msg.Body)
			default:
				log.Printf("main loop: un-handled "+
					"IRC output message type"+
					" '%d'", msg.Type)
			}
		}
	}
}
