// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This module allows for the registration and management of minions from IRC.
//

package main

import (
	"fmt"
	"strconv"
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

// Now is a wrapper around Now which always returns the same time in test mode.
func Now() time.Time {
	if cfg.TestMode {
		return time.Unix(1136239445, 0)
	}

	return time.Now()
}

// PrivMsg is the message handler for 'play' requests.  When the "ping" command
// is issued, a ping command is issued to all the minion with a unique
// timestamp.  This timestamp will be used to validated incoming ping
// responses.
func (module *PingModule) PrivMsg(msg *Message) {
	if len(module.PingStartTimes) > 0 {
		IRCPrivMsg(msg.ReplyTo, "error: previous ping still running")
		return
	}

	module.PingReplyTo = msg.ReplyTo

	for _, minion := range GetChannelMinions(msg.ReplyTo) {
		now := Now()
		module.PingStartTimes[minion.UserID] = now
		body := fmt.Sprintf("ping %d", now.UnixNano())
		SendToQueue(minion.QueueURL, body)
		Debug(fmt.Sprintf("sent to %s: %s", minion.Name, body))
	}

	// After 10 seconds, give up.
	go func() {
		time.Sleep(10 * time.Second)
		module.PingReset()
	}()
}

// PingReset resets all the ping internal variables.  This is used at the end
// of a ping to prepare the module for the next ping.
func (module *PingModule) PingReset() {
	module.PingStartTimes = make(map[string]time.Time)
	module.PingReplyTo = ""
}

// MinionMsg is the handler for minion responses.
func (module *PingModule) MinionMsg(msg *Message) {
	if len(msg.Args) != 1 {
		Debug("pong: usage error")
		return
	}

	timestamp, err := strconv.ParseInt(msg.Args[0], 10, 0)
	if err != nil {
		Debug("pong: invalid timestamp: " + err.Error())
		return
	}

	start, ok := module.PingStartTimes[msg.UserID]
	if !ok {
		Debug("pong: unknown minion: " + msg.UserID)
		return
	}
	delete(module.PingStartTimes, msg.UserID)

	if timestamp != start.UnixNano() {
		Debug(fmt.Sprintf("pong: got old ping reponse (%d)", timestamp))
		return
	}

	duration := time.Since(start)

	var name string
	minion, err := Minions.GetByUserID(msg.UserID)
	if err != nil {
		name = "no name (" + msg.UserID + ")"
	} else {
		name = minion.Name
	}

	reply := fmt.Sprintf("delay with %s: %s", name, duration)
	IRCPrivMsg(module.PingReplyTo, reply)
}

// Init registers all the commands for this module.
func (module *PingModule) Init() {
	module.PingReset()

	// ping/pong dance
	RegisterCommand(Command{
		Name:            "ping",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    true,
		AllowChannel:    true,
	})
	RegisterCommand(Command{
		Name:              "pong",
		MinionMsgFunction: module.MinionMsg,
		Addressed:         true,
		AllowPrivate:      true,
		AllowChannel:      true,
	})
}
