// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import(
    "strings"
)

// QueueModule controls the 'queue' command.
type QueueModule struct {
    *Server
}

// PrivMsg is the message handler for user 'queue' requests.
func (module *QueueModule) PrivMsg(srv *Server, msg *Message) {
    usage := "usage: q[ueue] url [s=start] [e=end]"
    track := "queueTrack"
    muted := "false"
    loop := "false"

    // validate command usage
    if len(msg.Args) < 1 || len(msg.Args) > 3 {
        srv.IRCPrivMsg(msg.ReplyTo, usage)
        return
    }

    // grab starting and ending time frame bounds if either is passed
    sBound, eBound, err := getBounds(msg.Args)
    if err != nil {
        srv.IRCPrivMsg(msg.ReplyTo, usage)
        return
    }

    // validate the passed value is a legitimate URI
    uri, err := parseURL(msg.Args[0])
    if err != nil {
        srv.IRCPrivMsg(msg.ReplyTo, err.Error())
        return
    }

    // if it's an imgur link, change any .giv/.gifv extension to a .webm
    if isImgur(uri) {
        uri, err = formatImgurURL(uri)
        if err != nil {
            srv.IRCPrivMsg(msg.ReplyTo, "error: couldn't format imgur URL")
            return
        }
    }

    mediaType := getMediaType(uri)
    srcValue := uri.String()
    switch mediaType {
    case "youtube":
        srcValue = reYTVideoId.FindAllStringSubmatch(uri.String(), -1)[0][2]
        break
    case "image", "web":
        errMsg := "error: URL must be of either an " +
            "audio file (" + strings.Join(audioFileExts, ", ") + "), " +
            "video file (" + strings.Join(videoFileExts, ", ") + "), " +
            "YouTube video, or imgur .gif/gifv."
        srv.IRCPrivMsg(msg.ReplyTo, errMsg)
        return
    }

    // send command to minions
    json := serializeMediaObj(track, mediaType, srcValue, sBound, eBound,
                              muted, loop)
    srv.SendToChannelMinions(msg.ReplyTo,
        "queue " + json)
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
