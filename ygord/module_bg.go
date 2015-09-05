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

    // validate command usage
    mediaList, err := parseArgList(msg.Args)
    if err != nil {
        srv.IRCPrivMsg(msg.ReplyTo, usage)
        return
    }

    mediaObjs := []*MediaObj{}

    for _, mediaItem := range mediaList {
        mObj := new (MediaObj)
        err := mObj.SetSrc(mediaItem["url"])
        if err != nil {
            srv.IRCPrivMsg(msg.ReplyTo, err.Error())
            return
        }

        if mObj.GetMediaType() == "audio" {
            errMsg := "error: backgrounds are seen, not heard (" +
                mObj.GetURL() + ")"
            srv.IRCPrivMsg(msg.ReplyTo, errMsg)
            return
        }

        mObj.Start = mediaItem["start"]
        mObj.End = mediaItem["end"]
        mObj.Muted = muted
        // construct the mediaObj for this mediaItem that will go into the
        // array in the media command JSON
        mediaObjs = append(mediaObjs, mObj)
    }

    // serialize the JSON that will be passed to the minions
    json := "{" +
                "\"status\":\"media\"," +
                "\"track\":\"" + track + "\"," +
                "\"loop\":" + loop + "," +
                "\"mediaObjs\":["
    for i, mObj := range mediaObjs {
        json += mObj.Serialize()
        if i < (len(mediaObjs) - 1){
            json += ","
        }
    }
    json +=     "]" +
            "}"

    // send command to minions
    srv.SendToChannelMinions(msg.ReplyTo,
        "bg " + json)
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
