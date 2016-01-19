// Copyright 2016, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"log"
	"math"
	"time"

	"github.com/truveris/ygor/ygord/alias"
)

// ScreensaverModule controls the 'image' command.
type ScreensaverModule struct {
	*Server
	Count uint32
}

// PrivMsg is the message handler for user 'image' requests.
func (module *ScreensaverModule) PrivMsg(srv *Server, msg *InputMessage) {
	srv.Reply(msg, "error: not implemented yet")
}

func (module *ScreensaverModule) StartScreensaver(srv *Server, client *Client, alias *alias.Alias) {
	log.Printf("screensaver: start client=%s alias=%s", client.ID, alias.Name)

	msgs, err := srv.NewMessagesFromBody(alias.Value, 0)
	if err != nil {
		log.Printf("screensaver: lexer/expand error: %s", err.Error)
		return
	}

	for _, msg := range msgs {
		msg.Type = InputMsgTypeScreensaver
		msg.Nickname = srv.Nickname
		msg.ReplyTo = client.ID
		srv.InputQueue <- msg
	}
}

// Tick runs every X seconds and checks for client screensaver delays.
func (module *ScreensaverModule) Tick(srv *Server) {
	// Every second since the beginning of time (1970) is put into a time
	// slot.  The division of these time slots is based on the configured
	// screensaver delay.  Which screensaver is started is based on that
	// slot of time.
	slot := time.Now().Unix() / int64(srv.Config.ScreensaverDelay)

	for _, client := range srv.ClientRegistry {
		age := time.Now().Sub(client.LastCommand)
		if age < time.Duration(srv.Config.ScreensaverDelay)*time.Second {
			continue
		}

		names := srv.Aliases.Find("screensaver/" + client.Channel + "/")
		if len(names) == 0 {
			continue
		}

		idx := int(math.Mod(float64(slot), float64(len(names))))

		alias := srv.Aliases.Get(names[idx])
		if alias == nil {
			continue
		}

		module.StartScreensaver(srv, client, alias)
	}
}

// Loop runs forever from the moment the module is initialized.  It waits
// either for a KeepAlive event or a timer to run out.
func (module *ScreensaverModule) Loop(srv *Server) {
	for {
		select {
		case <-time.After(5 * time.Second):
			module.Tick(srv)
		}
	}
}

// Init registers all the commands for this module.
func (module ScreensaverModule) Init(srv *Server) {
	if srv.Config.ScreensaverDelay > 0 {
		go module.Loop(srv)
	}

	srv.RegisterCommand(Command{
		Name:            "screensaver",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
