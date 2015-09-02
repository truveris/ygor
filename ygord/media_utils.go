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

// grabs starting and ending time frame bounds if either is passed
func getBounds(args []string) (string, string, error) {
    sBound := ""
    eBound := ""
    if len(args) > 1 {
        firstBound := strings.Split(args[1], "=")
        if len(firstBound) != 2 {
            return "", "", errors.New("invalid argument")
        }
        switch firstBound[0] {
        case "s":
            sBound = firstBound[1]
            break
        case "e":
            eBound = firstBound[1]
            break
        default:
            return "", "", errors.New("invalid argument")
        }
        if len(args) == 3 {
            secondBound := strings.Split(args[2], "=")
            if len(secondBound) != 2 {
                return "", "", errors.New("invalid argument")
            }
            switch secondBound[0] {
            case "s":
                sBound = secondBound[1]
                break
            case "e":
                eBound = secondBound[1]
                break
            default:
                return "", "", errors.New("invalid argument")
            }
        }
    }

    //everything's good
    return sBound, eBound, nil
}

func parseURL(link string) (*url.URL, error) {
    // validate the passed value is a legitimate URI
    uri, err := url.ParseRequestURI(link)
    if err != nil {
        errorMsg := "error: not a valid URL"
        return uri, errors.New(errorMsg)
    }

    // validate scheme is either HTTP, HTTPS, or FILE
    scheme := strings.ToUpper(uri.Scheme)
    if scheme != "HTTP" && scheme != "HTTPS" && scheme != "FILE" {
        errorMsg := "error: requires scheme of either HTTP, HTTPS, or FILE"
        return uri, errors.New(errorMsg)
    }

    if scheme != "FILE" {
        // validate hostname and port (if there is a port)
        hostParts := strings.Split(uri.Host, ":")
        if len(hostParts) > 2 {
            errorMsg := "error: invalid host"
            return uri, errors.New(errorMsg)
        } else if len(hostParts) == 2 {
            if !rePort.MatchString(hostParts[1]){
                errorMsg := "error: invalid port"
                return uri, errors.New(errorMsg)
            }
        }
        hostnameParts := strings.Split(hostParts[0], ".")
        // there needs to be at least 1 part
        if len(hostnameParts) < 1 {
            errorMsg := "error: invalid hostname"
            return uri, errors.New(errorMsg)
        }

        // validate the hostname parts
        for _, part := range hostnameParts {
            if !reHostnamePart.MatchString(part){
                errorMsg := "error: invalid hostname"
                return uri, errors.New(errorMsg)
            }
        }
    }

    // everything's good
    return uri, nil
}

// takes in a a url object and returns the media type
func getMediaType(uri *url.URL) string {
    // is it a youtube URI?
    for _, d := range youtubeHostNames {
        if uri.Host == d {
            return "youtube"
        }
    }

    // see if there's a file extension
    matches := reFileExt.FindAllStringSubmatch(uri.Path, -1)
    if len(matches) > 0 {
        file_ext := matches[0][1]

        // check if it's an image
        for _, ext := range imageFileExts {
            if file_ext == ext {
                return "image"
            }
        }

        // check if it's audio
        for _, ext := range audioFileExts {
            if file_ext == ext {
                return "audio"
            }
        }

        // check if it's video
        for _, ext := range videoFileExts {
            if file_ext == ext {
                return "video"
            }
        }
    }

    // if it isn't recognized as a supported file format, or a file extension
    // can't be found, just return 'web'
    return "web"
}

// takes in url object and returns a boolean based on whether or not it's an
// imgur domain
func isImgur(uri *url.URL) bool {
    for _, d := range imgurHostNames {
        if uri.Host == d {
            return true
        }
    }
    return false
}

// takes in url object and attempts to swap a gif or gifv extension with .webm
// and returns a url object made with the formatted URL
func formatImgurURL(uri *url.URL) (*url.URL, error) {
    formattedURL := reGifV.ReplaceAllString(uri.String(), ".webm")
    newUri, err := url.ParseRequestURI(formattedURL)
    return newUri, err
}

// returns stringified JSON mediaObj that is to be sent to minions
func serializeMediaObj(track string, mtype string, src string, s string,
                       e string, muted string, loop string) string {
    json := "{" +
                "\"status\":\"media\"," +
                "\"track\":\"" + track + "\"," +
                "\"mediaType\":\"" + mtype + "\"," +
                "\"src\":\"" + src + "\"," +
                "\"start\":\"" + s + "\"," +
                "\"end\":\"" + e + "\"," +
                "\"muted\":" + muted + "," +
                "\"loop\":" + loop +
            "}"
    return json
}
