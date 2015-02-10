// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"strings"
)

// TurretModule is the module handling all the turret commands.
type TurretModule struct{}

// PrivMsg is the message handler for user-received 'turret' commands.
func (module TurretModule) PrivMsg(msg *Message) {
	if len(msg.Args) == 0 || len(msg.Args) > 2 {
		IRCPrivMsg(msg.ReplyTo, "usage: turret command [param]")
		return
	}

	SendToChannelMinions(msg.ReplyTo, "turret "+strings.Join(msg.Args, " "))
}

// MinionMsg is the message handler for minion-received 'turret' commands.
func (module TurretModule) MinionMsg(msg *Message) {
	if msg.Args[0] != "ok" {
		minion, err := Minions.GetByUserID(msg.UserID)
		if err != nil {
			Debug(fmt.Sprintf("turret: can't find minion for %s",
				msg.UserID))
			return
		}
		channels := GetChannelsByMinionName(minion.Name)
		for _, channel := range channels {
			s := fmt.Sprintf("turret@%s: %s", minion.Name, strings.Join(msg.Args, " "))
			IRCPrivMsg(channel, s)
		}
	}
}

// Init registers all the commands for this module.
func (module TurretModule) Init() {
	RegisterCommand(Command{
		Name:              "turret",
		PrivMsgFunction:   module.PrivMsg,
		MinionMsgFunction: module.MinionMsg,
		Addressed:         true,
		AllowPrivate:      false,
		AllowChannel:      true,
	})
}
