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

	if cfg.IRCServer != "" {
		err = srv.StartIRCClient()
		if err != nil {
			log.Fatal("failed to start IRC client: ", err.Error())
		}
	}

	go waitForTraceRequest()

	log.Printf("ready, entering main loop")
	for {
		select {
		case msg := <-srv.InputQueue:
			log.Printf("irc %s <%s> %s", msg.ReplyTo,
				msg.Nickname, msg.Body)
			switch msg.Type {
			case InputMsgTypeIRCChannel:
				srv.IRCMessageHandler(msg)
			case InputMsgTypeIRCPrivate:
				srv.IRCMessageHandler(msg)
			case InputMsgTypeMattermost:
				srv.IRCMessageHandler(msg)
			default:
				log.Printf("main loop: un-handled "+
					"IRC input message type"+
					" '%d'", msg.Type)
			}
		case msg := <-srv.OutputQueue:
			log.Printf("irc %s <%s> %s", msg.Channel,
				cfg.Nickname, msg.Body)
			switch msg.Type {
			case OutputMsgTypePrivMsg:
				conn.Privmsg(msg.Channel, msg.Body)
			case OutputMsgTypeAction:
				conn.Action(msg.Channel, msg.Body)
			case OutputMsgTypeMattermost:
				srv.SendToMattermost(srv.NewMattermostResponse(msg.Channel,
					msg.Body))
			default:
				log.Printf("main loop: un-handled "+
					"IRC output message type"+
					" '%d'", msg.Type)
			}
		}
	}
}
