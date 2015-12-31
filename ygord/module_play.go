// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// PlayModule controls the 'play' command.
type PlayModule struct {
	*Server
}

// PrivMsg is the message handler for user 'play' requests.
func (module *PlayModule) PrivMsg(srv *Server, msg *InputMessage) {
	usage := "usage: play url [end]"

	// Validate the command's usage, and get back a map representing the media
	// item that was passed, along with it's start and end bounds.
	mediaItem, err := parseArgList(msg.Args)
	if err != nil {
		srv.Reply(msg, usage)
		return
	}

	media, err := NewMedia(srv, mediaItem, "playTrack", false, false,
		[]string{
			"soundcloud",
			"vimeo",
			"youtube",
			"video",
			"audio",
		})
	if err != nil {
		srv.Reply(msg, err.Error())
		return
	}

	// Send the command to the connected minions.
	srv.SendToChannelMinions(msg.ReplyTo, ClientCommand{"play", media})
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
