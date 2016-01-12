// Copyright 2015-2016, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
)

// ImageModule controls the 'image' command.
type ImageModule struct {
	*Server
}

// PrivMsg is the message handler for user 'image' requests.
func (module *ImageModule) PrivMsg(srv *Server, msg *InputMessage) {
	usage := "usage: image url [end]"

	// Validate the command's usage, and get back a map representing the media
	// item that was passed, along with it's start and end bounds.
	mediaItem, err := parseArgList(msg.Args)
	if err != nil {
		srv.Reply(msg, usage)
		return
	}

	media, err := NewMedia(srv, mediaItem, "imageTrack", true, true,
		[]string{
			"vimeo",
			"youtube",
			"video",
			"img",
			"web",
		})
	if err != nil {
		srv.Reply(msg, err.Error())
		return
	}

	// If a Mattermost message requests an image, it will be displayed in
	// the channel.  We also check the Depth to make sure we are not
	// displaying images right after Mattermost parsed a URL.  So if a user
	// calls ygor: image directly, that should be Depth 1, however if a
	// user uses an alias, that should be at least Depth 2 which will
	// trigger the following.
	if media.Format == "img" && msg.IsMattermost() && msg.Depth > 1 {
		srv.Reply(msg, fmt.Sprintf("![](%s)", media.Src))
	}

	// Send the command to the connected minions.
	srv.SendToChannelMinions(msg.ReplyTo, ClientCommand{"image", media})
}

// Init registers all the commands for this module.
func (module ImageModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:            "image",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
