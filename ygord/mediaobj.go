// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import(
    "errors"
    "net/url"
    "regexp"
    "strings"
)

var (
    // known domains to check for
    youtubeHostNames = []string{
        "www.youtube.com",
        "www.youtu.be",
        "youtube.com",
        "youtu.be",
    }
    imgurHostNames = []string{
        "i.imgur.com",
        "www.imgur.com",
        "imgur.com",
    }
    // known file extensions that are supported by Firefox
    audioFileExts = []string{
        "mp3",
        "wav",
        "wave",
    }
    imageFileExts = []string{
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
    videoFileExts = []string{
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

    // regex
    rePort = regexp.MustCompile(`^[0-9]+$`)
    reHostnamePart = regexp.MustCompile(`^([a-zA-Z0-9]+\-+)*[a-zA-Z0-9]+$`)
    reYTVideoId = regexp.MustCompile(
        `^.*(youtu.be\/|v\/|u\/\w\/|embed\/|watch\?v=|\&v=)([^#\&\?]*).*`)
    reGifV = regexp.MustCompile(`\.gif(v)?`)
    reFileExt = regexp.MustCompile(`.*\.([a-zA-Z0-9]+)[^a-zA-Z0-9]*$`)
)

type MediaObj struct {
    src       string // is formatted over time and will be passed to minions
    url       string // used to track the original URL passed by the command
    host      string
    path      string
    mediaType string
    Start     string
    End       string
    Muted     string
}

func (mObj *MediaObj) SetSrc(url string) error{
    uri, err := parseURL(url)
    if err != nil {
        errMsg := "error: " + err.Error() + " (" + url + ")"
        return errors.New(errMsg)
    }
    mObj.src = uri.String()
    mObj.url = url
    mObj.host = uri.Host
    mObj.path = uri.Path

    // if it's an imgur link, change any .giv/.gifv extension to a .webm
    if mObj.isImgur() {
        err := mObj.formatImgurURL()
        if err != nil {
            return err
        }
    }

    mObj.setMediaType()

    if mObj.isYouTube() {
        mObj.setYouTubeVideoID()
    }

    return nil
}

func (mObj *MediaObj) GetSrc() string {
    return mObj.src
}

func (mObj *MediaObj) GetURL() string {
    return mObj.url
}

func (mObj *MediaObj) setMediaType() {
    // is it a youtube URI?
    if mObj.isYouTube() {
        mObj.mediaType = "youtube"
        return
    }

    // see if there's a file extension
    matches := reFileExt.FindAllStringSubmatch(mObj.path, -1)
    if len(matches) > 0 {
        file_ext := matches[0][1]

        // check if it's an image
        for _, ext := range imageFileExts {
            if file_ext == ext {
                mObj.mediaType = "img"
                return
            }
        }

        // check if it's audio
        for _, ext := range audioFileExts {
            if file_ext == ext {
                mObj.mediaType = "audio"
                return
            }
        }

        // check if it's video
        for _, ext := range videoFileExts {
            if file_ext == ext {
                mObj.mediaType = "video"
                return
            }
        }
    }

    // if it isn't recognized as a supported file format, or a file extension
    // can't be found, just return 'web'
    mObj.mediaType = "web"
    return
}

func (mObj *MediaObj) GetMediaType() string {
    return mObj.mediaType
}

func (mObj *MediaObj) isImgur() bool {
    for _, d := range imgurHostNames {
        if mObj.host == d {
            return true
        }
    }
    return false
}

func (mObj *MediaObj) isYouTube() bool {
    for _, d := range youtubeHostNames {
        if mObj.host == d {
            return true
        }
    }
    return false
}

// swaps .gif/.gifv file extension of MediaObj's src with .webm
func (mObj *MediaObj) formatImgurURL() error {
    newUrl := reGifV.ReplaceAllString(mObj.src, ".webm")
    uri, err := parseURL(newUrl)
    if err != nil {
        errMsg := "error: " + err.Error() + " (" + mObj.url + ")"
        return errors.New(errMsg)
    }
    mObj.src = uri.String()
    mObj.path = uri.Path
    return nil
}

// sets the MediaObj's src as the YouTube video ID
func (mObj *MediaObj) setYouTubeVideoID() {
    mObj.src = reYTVideoId.FindAllStringSubmatch(mObj.src, -1)[0][2]
    return
}

func parseURL(link string) (*url.URL, error) {
    // validate the passed value is a legitimate URI
    uri, err := url.ParseRequestURI(link)
    if err != nil {
        errorMsg := "not a valid URL"
        return uri, errors.New(errorMsg)
    }

    // validate scheme is either HTTP, HTTPS, or FILE
    scheme := strings.ToUpper(uri.Scheme)
    if scheme != "HTTP" && scheme != "HTTPS" && scheme != "FILE" {
        errorMsg := "invalid scheme"
        return uri, errors.New(errorMsg)
    }

    if scheme != "FILE" {
        // validate hostname and port (if there is a port)
        hostParts := strings.Split(uri.Host, ":")
        if len(hostParts) > 2 {
            errorMsg := "invalid host"
            return uri, errors.New(errorMsg)
        } else if len(hostParts) == 2 {
            if !rePort.MatchString(hostParts[1]){
                errorMsg := "invalid port"
                return uri, errors.New(errorMsg)
            }
        }
        hostnameParts := strings.Split(hostParts[0], ".")
        // there needs to be at least 1 part
        if len(hostnameParts) < 1 {
            errorMsg := "invalid hostname"
            return uri, errors.New(errorMsg)
        }

        // validate the hostname parts
        for _, part := range hostnameParts {
            if !reHostnamePart.MatchString(part){
                errorMsg := "invalid hostname"
                return uri, errors.New(errorMsg)
            }
        }
    }

    // everything's good
    return uri, nil
}

// returns stringified JSON mediaObj that would normally be sent to minions
func (mObj *MediaObj) Serialize() string {
    json := "{" +
                "\"mediaType\":\"" + mObj.mediaType + "\"," +
                "\"src\":\"" + mObj.src + "\"," +
                "\"start\":\"" + mObj.Start + "\"," +
                "\"end\":\"" + mObj.End + "\"," +
                "\"muted\":" + mObj.Muted +
            "}"
    return json
}
