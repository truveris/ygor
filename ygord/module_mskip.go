// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// MSkipModule controls the 'mskip' command.
type MSkipModule struct{}

// PrivMsg is the message handler for user requests.
func (module *MSkipModule) PrivMsg(srv *Server, msg *Message) {
    srv.SendToChannelMinions(msg.ReplyTo, "mskip")
}

// Init registers all the commands for this module.
func (module MSkipModule) Init(srv *Server) {
    srv.RegisterCommand(Command{
        Name:            "mskip",
        PrivMsgFunction: module.PrivMsg,
        Addressed:       true,
        AllowPrivate:    false,
        AllowChannel:    true,
    })
}
