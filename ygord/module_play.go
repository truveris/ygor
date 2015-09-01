// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import(
    "regexp"
    "strings"
    "net/url"
    "log"
)

// PlayModule controls the 'play' command.
type PlayModule struct {
    *Server
}

// PrivMsg is the message handler for user 'play' requests.
func (module *PlayModule) PrivMsg(srv *Server, msg *Message) {
    // validate command usage
    if len(msg.Args) < 1 || len(msg.Args) > 3 {
        srv.IRCPrivMsg(msg.ReplyTo, "usage: play url [s=start] [e=end]")
        return
    }
    sBound := ""
    eBound := ""
    if len(msg.Args) > 1 {
        firstBound := strings.Split(msg.Args[1], "=")
        if len(firstBound) != 2 {
            srv.IRCPrivMsg(msg.ReplyTo, "usage: play url [s=start] [e=end]")
            return
        }
        switch {
        case firstBound[0] == "s":
            sBound = firstBound[1]
        case firstBound[0] == "e":
            eBound = firstBound[1]
        default:
            srv.IRCPrivMsg(msg.ReplyTo, "usage: play url [s=start] [e=end]")
            return
        }
        if len(msg.Args) == 3 {
            secondBound := strings.Split(msg.Args[2], "=")
            if len(secondBound) != 2 {
                srv.IRCPrivMsg(msg.ReplyTo, "usage: play url [s=start] [e=end]")
                return
            }
            switch {
            case secondBound[0] == "s":
                sBound = secondBound[1]
            case secondBound[0] == "e":
                eBound = secondBound[1]
            default:
                srv.IRCPrivMsg(msg.ReplyTo, "usage: play url [s=start] [e=end]")
                return
            }
        }
    }
    // known domains to check for
    youtubeHostNames := []string{
        "www.youtube.com",
        "www.youtu.be",
        "youtube.com",
        "youtu.be",
    }
    imgurHostNames := []string{
        "i.imgur.com",
        "www.imgur.com",
        "imgur.com",
    }
    // known file extensions that are supported by Firefox
    audioFileExts := []string{
        "mp3",
        "wav",
        "wave",
    }
    imageFileExts := []string{
        "jpg",
        "jpeg",
        "jpe",
        "jif",
        "jfif",
        "jfi",
        "png",
        "apng",
        "bmp",
        "dib",
        "gif",
    }
    videoFileExts := []string{
        "webm",
        "mp4",
        "m4a",
        "m4p",
        "m4b",
        "m4r",
        "m4v",
        "ogg",
        "ogv",
        "oga",
        "ogx",
        "ogm",
        "spx",
        "opus",
    }

    // validate the passed value is a legitimate URI
    uri, err := url.ParseRequestURI(msg.Args[0])
    log.Printf("%v", uri.Host)
    if err != nil {
        srv.IRCPrivMsg(msg.ReplyTo, "error: not a valid URI")
        return
    }

    // validate scheme is either HTTP, HTTPS, or FILE
    scheme := strings.ToUpper(uri.Scheme)
    if scheme != "HTTP" && scheme != "HTTPS" && scheme != "FILE" {
        srv.IRCPrivMsg(msg.ReplyTo, "error: requires scheme of either HTTP, HTTPS, or FILE")
        return
    }

    if scheme != "FILE" {
        // validate hostname and port (if there is a port)
        hostParts := strings.Split(uri.Host, ":")
        if len(hostParts) > 2 {
            srv.IRCPrivMsg(msg.ReplyTo, "error: invalid host")
            return
        } else if len(hostParts) == 2 {
            re := regexp.MustCompile(`^[0-9]+$`)
            if !re.MatchString(hostParts[1]){
                srv.IRCPrivMsg(msg.ReplyTo, "error: invalid port")
                return
            }
        }
        hostnameParts := strings.Split(hostParts[0], ".")
        log.Printf("%v", len(hostnameParts))
        // there needs to be at least 1 part
        if len(hostnameParts) < 1 {
            srv.IRCPrivMsg(msg.ReplyTo, "error: invalid hostname")
            return
        }

        //hostname parts can only include letters, numbers, and hyphens, but
        //hyphens can neither be the first, or last character of that part
        re := regexp.MustCompile(`^([a-zA-Z0-9]+\-+)*[a-zA-Z0-9]+$`)
        for _, part := range hostnameParts {
            if !re.MatchString(part){
                srv.IRCPrivMsg(msg.ReplyTo, "error: invalid hostname")
                return
            }
        }
    }

    //defaults
    mediaType := "web"
    srcValue := uri.String()

    // is it a youtube URI?
    for _, d := range youtubeHostNames {
        if uri.Host == d {
            mediaType = "youtube"
            re := regexp.MustCompile(`^.*(youtu.be\/|v\/|u\/\w\/|embed\/|watch\?v=|\&v=)([^#\&\?]*).*`)
            srcValue = re.FindAllStringSubmatch(uri.String(), -1)[0][2]
            break
        }
    }

    if mediaType != "youtube" {
        // if it's an imgur URI, replace .gif or .gifv with .webm
        for _, d := range imgurHostNames {
            if uri.Host == d {
                re := regexp.MustCompile(`\.gif(v)?`)
                uri, err = url.ParseRequestURI(re.ReplaceAllString(uri.String(), ".webm"))
                if err != nil {
                    srv.IRCPrivMsg(msg.ReplyTo, "error: couldn't format Imgur URL")
                    return
                }
                srcValue = uri.String()
                break
            }
        }

        // see if there's a file extension
        re := regexp.MustCompile(`.*\.([a-zA-Z0-9]+)[^a-zA-Z0-9]*$`)
        matches := re.FindAllStringSubmatch(uri.Path, -1)
        if len(matches) > 0 {
            file_ext := matches[0][1]
            //check if it's an image
            for _, ext := range imageFileExts {
                if file_ext == ext {
                    srv.IRCPrivMsg(msg.ReplyTo, "error: this command does not support images.")
                    return
                }
            }
            // check if it's audio
            for _, ext := range audioFileExts {
                if file_ext == ext {
                    mediaType = "audio"
                    break
                }
            }
            if mediaType != "audio" {
                // if it's not audio, check if it's a video
                for _, ext := range videoFileExts {
                    if file_ext == ext {
                        mediaType = "video"
                        break
                    }
                }
            }

            // if it's not a video/audio, get it outta here
            if mediaType != "audio" && mediaType != "video" {
                a := strings.Join(audioFileExts, ", ")
                v := strings.Join(videoFileExts, ", ")
                srv.IRCPrivMsg(msg.ReplyTo, "error: URL must be of either an audio file ("+a+"), video file("+v+"), or YouTube video.")
                return
            }
        } else {
            srv.IRCPrivMsg(msg.ReplyTo, "error: no file extension could be found")
            return
        }
    }

    // send command to minions
    json := "{" +
                "\"status\":\"media\"," +
                "\"track\":\"playTrack\"," +
                "\"mediaType\":\"" + mediaType + "\"," +
                "\"src\":\"" + srcValue + "\"," +
                "\"start\":\"" + sBound + "\"," +
                "\"end\":\"" + eBound + "\"," +
                "\"muted\":false," +
                "\"loop\":false" +
            "}"
    srv.SendToChannelMinions(msg.ReplyTo,
        "play "+json)
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
