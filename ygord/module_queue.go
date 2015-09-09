// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// QueueModule controls the 'queue' command.
type QueueModule struct {
	*Server
}

// PrivMsg is the message handler for user 'queue' requests.
func (module *QueueModule) PrivMsg(srv *Server, msg *Message) {
	usage := "usage: q[ueue] url [s=start] [e=end][, url [s=start] [e=end]]..."
	track := "queueTrack"
	muted := "false"
	loop := "false"

	// Validate the command's usage, and get back a map array representing the
	// media items that were passed, along with each one's start and end
	// bounds.
	mediaList, err := parseArgList(msg.Args)
	if err != nil {
		srv.IRCPrivMsg(msg.ReplyTo, usage)
		return
	}

	// Make the array that will house the pointers to the MediaObjs.
	mediaObjs := []*MediaObj{}

	// Parse all the media items in 'mediaList' into MediaObjs
	for _, mediaItem := range mediaList {
		mObj := new(MediaObj)
		err := mObj.SetSrc(mediaItem["url"])
		if err != nil {
			srv.IRCPrivMsg(msg.ReplyTo, err.Error())
			return
		}

		// Check the media types.
		if mObj.GetMediaType() == "img" || mObj.GetMediaType() == "web" {
			errMsg := "error: URL must be audio file, video file, YouTube " +
				"video, or imgur .gif/gifv. (" + mObj.GetURL() + ")"
			srv.IRCPrivMsg(msg.ReplyTo, errMsg)
			return
		}

		mObj.Start = mediaItem["start"]
		mObj.End = mediaItem["end"]
		mObj.Muted = muted
		// Append the constructed MediaObj onto the mediaObjs array.
		mediaObjs = append(mediaObjs, mObj)
	}

	// Serialize the JSON that will be passed to the connected minions.
	json := "{" +
		"\"status\":\"media\"," +
		"\"track\":\"" + track + "\"," +
		"\"loop\":" + loop + "," +
		"\"mediaObjs\":["
	for i, mObj := range mediaObjs {
		json += mObj.Serialize()
		if i < (len(mediaObjs) - 1) {
			// Add a comma after each MediaObj, unless it's the last one.
			json += ","
		}
	}
	json += "]" +
		"}"

	// Send the command to the connected minions.
	srv.SendToChannelMinions(msg.ReplyTo,
		"queue "+json)
}

// Init registers all the commands for this module.
func (module QueueModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:            "queue",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	// shorthand for 'queue' command
	srv.RegisterCommand(Command{
		Name:            "q",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
