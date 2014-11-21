// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/truveris/ygor"
)

var (
	rePercentage = regexp.MustCompile(`^[-+]?\d+%$`)
)

type VolumeModule struct{}

// Send the volume command to the minions.
func (module VolumeModule) PrivMsg(msg *ygor.Message) {
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

func (module VolumeModule) MinionMsg(msg *ygor.Message) {
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

func (module VolumeModule) Init() {
	ygor.RegisterCommand(ygor.Command{
		Name:              "volume",
		PrivMsgFunction:   module.PrivMsg,
		MinionMsgFunction: module.MinionMsg,
		Addressed:         true,
		AllowPrivate:      false,
		AllowChannel:      true,
	})
}
