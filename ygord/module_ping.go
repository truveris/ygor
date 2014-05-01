// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This module allows for the registration and management of minions from IRC.
//

package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/truveris/ygor"
)

// Basic module.
type PingModule struct {
	// When this is zero, there is no ping going on. If a ping is running,
	// the first value contains the start time of the ping and second the
	// channel/user to reply to.
	PingStartTimes map[string]time.Time

	// FIXME: once we have a request/response mechanism, this should be
	// deleted.
	PingReplyTo string
}

func (module *PingModule) PrivMsg(msg *ygor.Message) {}

// Wrapper around Now which always returns the same time in test mode.
func Now() time.Time {
	if cfg.TestMode {
		return time.Unix(1136239445, 0)
	}

	return time.Now()
}

// When the "ping" command is issued, send a ping command to all the minion
// with a unique timestamp. This timestamp will be used to validated incoming
// ping responses.
func (module *PingModule) PingCmdFunc(msg *ygor.Message) {
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

// Reset all the ping internal variables.
func (module *PingModule) PingReset() {
	module.PingStartTimes = make(map[string]time.Time)
	module.PingReplyTo = ""
}

// Minions reponding to ping.
func (module *PingModule) PongCmdFunc(msg *ygor.Message) {
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

func (module *PingModule) Init() {
	module.PingReset()

	// ping/pong dance
	ygor.RegisterCommand(ygor.Command{
		Name:            "ping",
		PrivMsgFunction: module.PingCmdFunc,
		Addressed:       true,
		AllowPrivate:    true,
		AllowChannel:    true,
	})
	ygor.RegisterCommand(ygor.Command{
		Name:              "pong",
		MinionMsgFunction: module.PongCmdFunc,
		Addressed:         true,
		AllowPrivate:      true,
		AllowChannel:      true,
	})
}
