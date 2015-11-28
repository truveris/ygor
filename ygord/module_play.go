// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// PlayModule controls the 'play' command.
type PlayModule struct {
	*Server
}

// PrivMsg is the message handler for user 'play' requests.
func (module *PlayModule) PrivMsg(srv *Server, msg *Message) {
	usage := "usage: play url [end]"

	// Validate the command's usage, and get back a map representing the media
	// item that was passed, along with it's start and end bounds.
	mediaItem, parseArgErr := parseArgList(msg.Args)
	if parseArgErr != nil {
		srv.IRCPrivMsg(msg.ReplyTo, usage)
		return
	}

	mObj, parseMObjErr := NewMediaObj(mediaItem, "playTrack", false, false,
		[]string{
			"youtube",
			"video",
			"audio",
		})
	if parseMObjErr != nil {
		srv.IRCPrivMsg(msg.ReplyTo, parseMObjErr.Error())
		return
	}

	// Send the command to the connected minions.
	srv.SendToChannelMinions(msg.ReplyTo,
		"play "+mObj.Serialize())
}

// Init registers all the commands for this module.
func (module PlayModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:            "play",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
