// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// BgModule controls the 'bg' command.
type BgModule struct {
	*Server
}

// PrivMsg is the message handler for user 'bg' requests.
func (module *BgModule) PrivMsg(srv *Server, msg *Message) {
	usage := "usage: bg url [s=start] [e=end][, url [s=start] [e=end]]..."
	track := "bgTrack"
	muted := "true"
	loop := "true"

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

		// The bgTrack shouldn't use audio, because it's meant for visuals and
		// everything it shows should be muted anyway.
		if mObj.GetMediaType() == "audio" {
			errMsg := "error: backgrounds are seen, not heard (" +
				mObj.GetURL() + ")"
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
		"bg "+json)
}

// Init registers all the commands for this module.
func (module BgModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:            "bg",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
