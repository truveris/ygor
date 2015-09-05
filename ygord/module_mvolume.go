// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
    "fmt"
    "log"
    "strings"
)

// var (
//     rePercentage = regexp.MustCompile(`^\d{1,3}%$`)
// )

// MVolumeModule is the module handling all the mvolume related commands.
type MVolumeModule struct{}

// PrivMsg is the message handler for user 'mvolume' requests.
func (module MVolumeModule) PrivMsg(srv *Server, msg *Message) {
    if len(msg.Args) != 1 {
        srv.IRCPrivMsg(msg.ReplyTo, "usage: mvolume percent")
        return
    }

    if !rePercentage.MatchString(msg.Args[0]) {
        srv.IRCPrivMsg(msg.ReplyTo, "error: bad input, must be absolute rounded percent value (e.g. 42%)")
        return
    }

    srv.SendToChannelMinions(msg.ReplyTo, "mvolume "+msg.Args[0])
}

// PrivMsgPlusPlus is the message handler for user 'mvolume++' requests, it
// increments the mvolume by 1dB.
func (module MVolumeModule) PrivMsgPlusPlus(srv *Server, msg *Message) {
    if len(msg.Args) != 0 {
        srv.IRCPrivMsg(msg.ReplyTo, "usage: mvolume++")
        return
    }
    srv.SendToChannelMinions(msg.ReplyTo, "mvolume 1dB+")
}

// PrivMsgMinusMinus is the message handler for user 'mvolume--' requests, it
// decrements the mvolume by 1dB.
func (module MVolumeModule) PrivMsgMinusMinus(srv *Server, msg *Message) {
    if len(msg.Args) != 0 {
        srv.IRCPrivMsg(msg.ReplyTo, "usage: mvolume--")
        return
    }
    srv.SendToChannelMinions(msg.ReplyTo, "mvolume 1dB-")
}

// MinionMsg is the message handler for all the minion responses for 'mvolume'
// requests.
func (module MVolumeModule) MinionMsg(srv *Server, msg *Message) {
    if msg.Args[0] != "ok" {
        minion, err := srv.Minions.GetByUserID(msg.UserID)
        if err != nil {
            log.Printf("mvolume: can't find minion for %s", msg.UserID)
            return
        }
        channels := srv.Config.GetChannelsByMinion(minion.Name)
        for _, channel := range channels {
            s := fmt.Sprintf("mvolume@%s: %s", minion.Name, strings.Join(msg.Args, " "))
            srv.IRCPrivMsg(channel, s)
        }
    }
}

// Init registers all the commands for this module.
func (module MVolumeModule) Init(srv *Server) {
    srv.RegisterCommand(Command{
        Name:              "mvolume",
        PrivMsgFunction:   module.PrivMsg,
        MinionMsgFunction: module.MinionMsg,
        Addressed:         true,
        AllowPrivate:      false,
        AllowChannel:      true,
    })

    srv.RegisterCommand(Command{
        Name:            "mvolume++",
        PrivMsgFunction: module.PrivMsgPlusPlus,
        Addressed:       true,
        AllowPrivate:    false,
        AllowChannel:    true,
    })

    srv.RegisterCommand(Command{
        Name:            "mvolume--",
        PrivMsgFunction: module.PrivMsgMinusMinus,
        Addressed:       true,
        AllowPrivate:    false,
        AllowChannel:    true,
    })
}
