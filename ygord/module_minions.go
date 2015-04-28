// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This module allows for the registration and management of minions from IRC.
//

package main

import (
	"log"
	"sort"
	"strings"
)

// MinionsModule controls the registration process for minions via the
// 'register' minion command.
type MinionsModule struct {
	*Server
}

// PrivMsg is the message handler for user 'minions' requests.
func (module *MinionsModule) PrivMsg(srv *Server, msg *Message) {
	var names []string

	minions, err := srv.Minions.All()
	if err != nil {
		log.Printf("GetMinions error: %s", err.Error())
		return
	}

	for _, minion := range minions {
		names = append(names, minion.Name)
	}
	sort.Strings(names)
	srv.IRCPrivMsg(msg.ReplyTo,
		"currently registered: "+strings.Join(names, ", "))
}

// MinionMsg is the message handler for minions 'register' requests.
func (module *MinionsModule) MinionMsg(srv *Server, msg *Message) {
	if len(msg.Args) != 2 {
		log.Printf("register: error: invalid register command issued")
		return
	}

	name := msg.Args[0]
	queueURL := msg.Args[1]

	err := srv.Minions.Register(name, queueURL, msg.UserID)
	if err != nil {
		log.Printf("register: error: %s", err.Error())
		return
	}

	srv.Minions.Save()
	srv.SendToQueue(queueURL, "register success")
}

// Init registers all the commands for this module.
func (module *MinionsModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:              "register",
		MinionMsgFunction: module.MinionMsg,
		Addressed:         true,
		AllowPrivate:      true,
		AllowChannel:      true,
	})

	srv.RegisterCommand(Command{
		Name:            "minions",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    true,
		AllowChannel:    true,
	})
}
