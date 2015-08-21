// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// MediaModule controls the 'media' command.
type MediaModule struct {
    *Server
}

// PrivMsg is the message handler for user 'media' requests.
func (module *MediaModule) PrivMsg(srv *Server, msg *Message) {
    if len(msg.Args) != 1 {
        srv.IRCPrivMsg(msg.ReplyTo, "usage: media url")
        return
    }

    srv.SendToChannelMinions(msg.ReplyTo,
        "xombrero open /fullscreen.html?"+msg.Args[0])
}

// Init registers all the commands for this module.
func (module MediaModule) Init(srv *Server) {
    srv.RegisterCommand(Command{
        Name:            "media",
        PrivMsgFunction: module.PrivMsg,
        Addressed:       true,
        AllowPrivate:    false,
        AllowChannel:    true,
    })
}
