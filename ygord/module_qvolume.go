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

// QVolumeModule is the module handling all the qvolume related commands.
type QVolumeModule struct{}

// PrivMsg is the message handler for user 'qvolume' requests.
func (module QVolumeModule) PrivMsg(srv *Server, msg *Message) {
    if len(msg.Args) != 1 {
        srv.IRCPrivMsg(msg.ReplyTo, "usage: qvolume percent")
        return
    }

    if !rePercentage.MatchString(msg.Args[0]) {
        srv.IRCPrivMsg(msg.ReplyTo, "error: bad input, must be absolute rounded percent value (e.g. 42%)")
        return
    }

    srv.SendToChannelMinions(msg.ReplyTo, "qvolume "+msg.Args[0])
}

// PrivMsgPlusPlus is the message handler for user 'qvolume++' requests, it
// increments the qvolume by 1dB.
func (module QVolumeModule) PrivMsgPlusPlus(srv *Server, msg *Message) {
    if len(msg.Args) != 0 {
        srv.IRCPrivMsg(msg.ReplyTo, "usage: qvolume++")
        return
    }
    srv.SendToChannelMinions(msg.ReplyTo, "qvolume 1dB+")
}

// PrivMsgMinusMinus is the message handler for user 'qvolume--' requests, it
// decrements the qvolume by 1dB.
func (module QVolumeModule) PrivMsgMinusMinus(srv *Server, msg *Message) {
    if len(msg.Args) != 0 {
        srv.IRCPrivMsg(msg.ReplyTo, "usage: qvolume--")
        return
    }
    srv.SendToChannelMinions(msg.ReplyTo, "qvolume 1dB-")
}

// MinionMsg is the message handler for all the minion responses for 'qvolume'
// requests.
func (module QVolumeModule) MinionMsg(srv *Server, msg *Message) {
    if msg.Args[0] != "ok" {
        minion, err := srv.Minions.GetByUserID(msg.UserID)
        if err != nil {
            log.Printf("qvolume: can't find minion for %s", msg.UserID)
            return
        }
        channels := srv.Config.GetChannelsByMinion(minion.Name)
        for _, channel := range channels {
            s := fmt.Sprintf("qvolume@%s: %s", minion.Name, strings.Join(msg.Args, " "))
            srv.IRCPrivMsg(channel, s)
        }
    }
}

// Init registers all the commands for this module.
func (module QVolumeModule) Init(srv *Server) {
    srv.RegisterCommand(Command{
        Name:              "qvolume",
        PrivMsgFunction:   module.PrivMsg,
        MinionMsgFunction: module.MinionMsg,
        Addressed:         true,
        AllowPrivate:      false,
        AllowChannel:      true,
    })

    srv.RegisterCommand(Command{
        Name:            "qvolume++",
        PrivMsgFunction: module.PrivMsgPlusPlus,
        Addressed:       true,
        AllowPrivate:    false,
        AllowChannel:    true,
    })

    srv.RegisterCommand(Command{
        Name:            "qvolume--",
        PrivMsgFunction: module.PrivMsgMinusMinus,
        Addressed:       true,
        AllowPrivate:    false,
        AllowChannel:    true,
    })
}
