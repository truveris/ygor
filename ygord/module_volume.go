// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	rePercentage = regexp.MustCompile(`^[-+]?\d+%$`)
)

// VolumeModule is the module handling all the volume related commands.
type VolumeModule struct{}

// PrivMsg is the message handler for user 'volume' requests.
func (module VolumeModule) PrivMsg(msg *Message) {
	if len(msg.Args) != 1 {
		IRCPrivMsg(msg.ReplyTo, "usage: volume percent")
		return
	}

	if !rePercentage.MatchString(msg.Args[0]) {
		IRCPrivMsg(msg.ReplyTo, "error: bad input, must be percent")
		return
	}

	SendToChannelMinions(msg.ReplyTo, "volume "+msg.Args[0])
}

// MinionMsg is the message handler for all the minion responses for 'volume'
// requests.
func (module VolumeModule) MinionMsg(msg *Message) {
	if msg.Args[0] != "ok" {
		minion, err := Minions.GetByUserID(msg.UserID)
		if err != nil {
			Debug(fmt.Sprintf("volume: can't find minion for %s",
				msg.UserID))
			return
		}
		channels := GetChannelsByMinionName(minion.Name)
		for _, channel := range channels {
			s := fmt.Sprintf("volume@%s: %s", minion.Name, strings.Join(msg.Args, " "))
			IRCPrivMsg(channel, s)
		}
	}
}

// Init registers all the commands for this module.
func (module VolumeModule) Init() {
	RegisterCommand(Command{
		Name:              "volume",
		PrivMsgFunction:   module.PrivMsg,
		MinionMsgFunction: module.MinionMsg,
		Addressed:         true,
		AllowPrivate:      false,
		AllowChannel:      true,
	})
}
