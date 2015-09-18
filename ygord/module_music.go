// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// MusicModule controls the 'music' command.
type MusicModule struct {
	*Server
}

// PrivMsg is the message handler for user 'music' requests.
func (module *MusicModule) PrivMsg(srv *Server, msg *Message) {
	usage := "usage: music url [-s start] [-e end][, url [-s start] [-e end]]..."
	track := "musicTrack"
	muted := false
	loop := false
	acceptableMediaTypes := []string{
		"video",
		"youtube",
		"audio",
		"soundcloud",
		"vimeo",
	}

	// Validate the command's usage, and get back a map array representing the
	// media items that were passed, along with each one's start and end
	// bounds.
	mediaList, err := parseArgList(msg.Args)
	if err != nil {
		srv.IRCPrivMsg(msg.ReplyTo, usage)
		return
	}

	// Make the MediaObjList that will house the pointers to the MediaObjs.
	mObjList := &MediaObjList{
		Track: track,
		Loop:  loop,
	}

	// Parse all the media items in 'mediaList' into MediaObjs
	for _, mediaItem := range mediaList {
		mObj := new(MediaObj)
		err := mObj.SetSrc(mediaItem["url"])
		if err != nil {
			srv.IRCPrivMsg(msg.ReplyTo, err.Error())
			return
		}

		// The musicTrack shouldn't embed images or webpages, because they are
		// assumed to not have audio, so there wouldn't be a point in embedding
		// them. Also, they don't end, so musicTrack won't ever move to the
		// next item in the queue.
		if !mObj.IsOfMediaType(acceptableMediaTypes) {
			errMsg := "error: music is heard, not seen (" + mObj.GetURL() + ")"
			srv.IRCPrivMsg(msg.ReplyTo, errMsg)
			return
		}

		mObj.Start = mediaItem["start"]
		mObj.End = mediaItem["end"]
		mObj.Muted = muted
		// Add the constructed MediaObj to the MediaObjList.
		mObjList.Append(mObj)
	}

	// Serialize the JSON that will be passed to the connected minions.
	json := mObjList.Serialize()

	// Send the command to the connected minions.
	srv.SendToChannelMinions(msg.ReplyTo,
		"music "+json)
}

// Init registers all the commands for this module.
func (module MusicModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:            "music",
		PrivMsgFunction: module.PrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})
}
