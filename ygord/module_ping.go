// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This module allows for the registration and management of minions from IRC.
//

package main

import (
	"time"
)

// PingModule controls the 'ping' / 'pong' commands.
type PingModule struct {
	// PingStartTime keeps track of the ping request start times on a
	// per-minion basis.  The map key is the minion id and the time is the
	// moment when the request was made.
	PingStartTimes map[string]time.Time

	// PingReplyTo sets who to send the results to.
	PingReplyTo string
}

// PrivMsg is the message handler for 'play' requests.  When the "ping" command
// is issued, a ping command is issued to all the minion with a unique
// timestamp.  This timestamp will be used to validated incoming ping
// responses.
func (module *PingModule) PrivMsg(srv *Server, msg *InputMessage) {
	srv.Reply(msg, "error: not implemented")
	return

	//	if len(module.PingStartTimes) > 0 {
	//		srv.IRCPrivMsg(msg.ReplyTo, "error: previous ping still running")
	//		return
	//	}
	//
	//	module.PingReplyTo = msg.ReplyTo
	//
	//	for _, minion := range srv.GetMinionsByChannel(msg.ReplyTo) {
	//		now := time.Now()
	//		module.PingStartTimes[minion.UserID] = now
	//		body := fmt.Sprintf("ping %d", now.UnixNano())
	//		srv.SendToQueue(minion.QueueURL, body)
	//		log.Printf("sent to %s: %s", minion.Name, body)
	//	}
	//
	//	// After 10 seconds, give up.
	//	go func() {
	//		time.Sleep(10 * time.Second)
	//		module.PingReset()
	//	}()
}

// PingReset resets all the ping internal variables.  This is used at the end
// of a ping to prepare the module for the next ping.
func (module *PingModule) PingReset() {
	module.PingStartTimes = make(map[string]time.Time)
	module.PingReplyTo = ""
}

// MinionMsg is the handler for minion responses.
func (module *PingModule) MinionMsg(srv *Server, msg *InputMessage) {
	return
	//	if len(msg.Args) != 1 {
	//		log.Printf("pong: usage error")
	//		return
	//	}
	//
	//	timestamp, err := strconv.ParseInt(msg.Args[0], 10, 0)
	//	if err != nil {
	//		log.Printf("pong: invalid timestamp: %s", err.Error())
	//		return
	//	}
	//
	//	start, ok := module.PingStartTimes[msg.UserID]
	//	if !ok {
	//		log.Printf("pong: unknown minion: %s", msg.UserID)
	//		return
	//	}
	//	delete(module.PingStartTimes, msg.UserID)
	//
	//	if timestamp != start.UnixNano() {
	//		log.Printf("pong: got old ping response (%d)", timestamp)
	//		return
	//	}
	//
	//	duration := time.Since(start)
	//
	//	var name string
	//	minion, err := srv.Minions.GetByUserID(msg.UserID)
	//	if err != nil {
	//		name = "no name (" + msg.UserID + ")"
	//	} else {
	//		name = minion.Name
	//	}
	//
	//	reply := fmt.Sprintf("delay with %s: %s", name, duration)
	//	srv.IRCPrivMsg(module.PingReplyTo, reply)
}

// Init registers all the commands for this module.
func (module *PingModule) Init(srv *Server) {
	module.PingReset()

	// ping/pong dance
	srv.RegisterCommand(Command{
		Name:            "ping",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    true,
		AllowChannel:    true,
	})
	// srv.RegisterCommand(Command{
	// 	Name:              "pong",
	// 	MinionMsgFunction: module.MinionMsg,
	// 	Addressed:         true,
	// 	AllowPrivate:      true,
	// 	AllowChannel:      true,
	// })
}
