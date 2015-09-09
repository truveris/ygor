// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

var (
	// These are the known domains to check for, where special formatting of
	// the passed URL is required so connected minions can most effectively
	// embed and manipulate the desired content.
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

	// These are the known file extensions that are supported by Firefox.
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

	rePort         = regexp.MustCompile(`^[0-9]+$`)
	reHostnamePart = regexp.MustCompile(`^([a-zA-Z0-9]+\-+)*[a-zA-Z0-9]+$`)
	reYTVideoID    = regexp.MustCompile(
		`^.*(youtu.be\/|v\/|u\/\w\/|embed\/|watch\?v=|\&v=)([^#\&\?]*).*`)
	reGifV    = regexp.MustCompile(`\.gif(v)?`)
	reFileExt = regexp.MustCompile(`.*\.([a-zA-Z0-9]+)[^a-zA-Z0-9]*$`)
)

// MediaObj represents the relevant data that will eventually be passed to
// the connected minions. It is used to generate the information that connected
// minions would use to properly embed the desired content.
//
// It also provides several functions that can be used to more easily work with
// the data, so that command modules aren't filled with a lot of excessive
// code.
type MediaObj struct {
	// 'src' is formatted over time and is what will eventually be passed to
	// the connected minions.
	src       string
	url       string
	host      string
	path      string
	// 'mediaType' tells the connected minions how to embed the desired content
	// using 'src'.
	mediaType string
	// Start represents where in the desired content's timeline to begin
	// playing
	Start     string
	// End represents where in the desired content's timeline to stop playing
	End       string
	// Muted represents whether or not the desired content should be muted
	Muted     string
}

// SetSrc takes in a string that represents a URL. This function determines if
// the URL is a valid URL, formats imgur links to use .webm instead of .gif(v),
// determines the mediaType that the URL represents, and grabs the videoID from
// YouTube links.
//
// The MediaObj's 'src' attribute will either be set to the passed URL, the
// formatted imgur URL (if it was an imgur link), or the YouTube video's
// videoID (if it was a YouTube video).
//
// The MediaObj's 'src' attribute can be retrieved using the MediaObj's
// 'GetSrc()' function.
//
// The URL that was originally passed, is saved as the MediaObj's 'url'
// attribute, and can be retrieved with the MediaObj's 'GetURL()' function.
func (mObj *MediaObj) SetSrc(url string) error {
	uri, err := parseURL(url)
	if err != nil {
		errMsg := "error: " + err.Error() + " (" + url + ")"
		return errors.New(errMsg)
	}
	mObj.src = uri.String()
	mObj.url = url
	mObj.host = uri.Host
	mObj.path = uri.Path

	// If it's an imgur link, change any .giv(v) extension to .webm, so minions
	// can embed it as a video, instead of an image.
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

// GetSrc returns the MediaObj's 'src' attribute (this is what should get passed to
// the connected minions).
func (mObj *MediaObj) GetSrc() string {
	return mObj.src
}

// GetURL returns the URL that was originally passed to the 'SetSrc()' function.
func (mObj *MediaObj) GetURL() string {
	return mObj.url
}

// setMediaType sets the 'mediaType' attribute of the MediaObj. This tells the
// connected minions what kind of content they should be trying to embed.
func (mObj *MediaObj) setMediaType() {
	// Is the passed URL a YouTube video link?
	if mObj.isYouTube() {
		mObj.mediaType = "youtube"
		return
	}

	// Does the passed URL have a file extension that can be used to determine
	// mediaType?
	matches := reFileExt.FindAllStringSubmatch(mObj.path, -1)
	if len(matches) > 0 {
		fileExt := matches[0][1]

		// Is it an image file?
		for _, ext := range imageFileExts {
			if fileExt == ext {
				mObj.mediaType = "img"
				return
			}
		}

		// Is it an audio file?
		for _, ext := range audioFileExts {
			if fileExt == ext {
				mObj.mediaType = "audio"
				return
			}
		}

		// Is it a video file?
		for _, ext := range videoFileExts {
			if fileExt == ext {
				mObj.mediaType = "video"
				return
			}
		}
	}

	// If it's not a link to a YouTube video, and it's not a link to an image
	// file, audio file, or video file (at least, of those that are supported
	// by Firefox), then make the MediaObj's 'mediaType' be 'web' so the
	// connected minions can just embed the URL as an iframe and hope for the
	// best.
	mObj.mediaType = "web"
	return
}

// GetMediaType returnes the MediaObj's 'mediaType' attribute. The 'mediaType'
// tells the connected minions what kind of content they should be trying to
// embed when using the MediaObj's 'src' attribute.
func (mObj *MediaObj) GetMediaType() string {
	return mObj.mediaType
}

// isImgur attempts to determine if the desired content is hosted on imgur
func (mObj *MediaObj) isImgur() bool {
	for _, d := range imgurHostNames {
		if mObj.host == d {
			return true
		}
	}
	return false
}

// isYouTube attempts to determine if the desired content is a video hosted on
// YouTube
func (mObj *MediaObj) isYouTube() bool {
	for _, d := range youtubeHostNames {
		if mObj.host == d {
			return true
		}
	}
	return false
}

// formatImgurURL swaps .gif(v) file extension of MediaObj's src with .webm
//
// imgur will automatically convert any .gif to .webm, but wrap it as a .gifv.
// WEBM files take far less bandwidth to download than their .gif counterparts,
// and are much easier to render, as well.
//
// Formatting the URL to use .webm (if applicable), allows the URL to be
// recognized as a video file (and thus, will make the 'mediaType' be 'video'),
// so when passed to the connected minions, they can embed it as a video,
// rather than an image.
func (mObj *MediaObj) formatImgurURL() error {
	newURL := reGifV.ReplaceAllString(mObj.src, ".webm")
	uri, err := parseURL(newURL)
	if err != nil {
		errMsg := "error: " + err.Error() + " (" + mObj.url + ")"
		return errors.New(errMsg)
	}
	mObj.src = uri.String()
	mObj.path = uri.Path
	return nil
}

// setYouTubeVideoID grabs the YouTube video's videoID from the passed URL, and
// sets it as the MediaObj's 'src' attribute.
func (mObj *MediaObj) setYouTubeVideoID() {
	mObj.src = reYTVideoID.FindAllStringSubmatch(mObj.src, -1)[0][2]
	return
}

// parseURL determines if the passed URL is acceptable, and then (if it's
// valid) returns a pointer to the URL object that was made using the passed
// URL.
func parseURL(link string) (*url.URL, error) {
	// validate the passed value is a legitimate URI
	uri, err := url.ParseRequestURI(link)
	if err != nil {
		errorMsg := "not a valid URL"
		return uri, errors.New(errorMsg)
	}

	// Validate that the scheme is either HTTP, HTTPS, or FILE.
	scheme := strings.ToUpper(uri.Scheme)
	if scheme != "HTTP" && scheme != "HTTPS" && scheme != "FILE" {
		errorMsg := "invalid scheme"
		return uri, errors.New(errorMsg)
	}

	if scheme != "FILE" {
		// Validate that the hostname and port (if there is a port) are
		// acceptible.
		hostParts := strings.Split(uri.Host, ":")
		if len(hostParts) > 2 {
			errorMsg := "invalid host"
			return uri, errors.New(errorMsg)
		} else if len(hostParts) == 2 {
			if !rePort.MatchString(hostParts[1]) {
				errorMsg := "invalid port"
				return uri, errors.New(errorMsg)
			}
		}
		hostnameParts := strings.Split(hostParts[0], ".")
		// Validate that there is at least something to the hostname.
		if len(hostnameParts) < 1 {
			errorMsg := "invalid hostname"
			return uri, errors.New(errorMsg)
		}

		// Validate that the hostname parts are limited to letters, numbers,
		// and dashes, and that the dashes are neither the first nor last
		// character in any of the parts.
		for _, part := range hostnameParts {
			if !reHostnamePart.MatchString(part) {
				errorMsg := "invalid hostname"
				return uri, errors.New(errorMsg)
			}
		}
	}

	// If it made it to this point, it must be a valid URL, so return the
	// pointer to the URL object.
	return uri, nil
}

// Serialize returns stringified JSON representation of the MediaObj. This is
// what would normally be passed to the connected minions.
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
