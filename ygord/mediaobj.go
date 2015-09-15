// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"encoding/json"
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
	// 'Src' is formatted over time and is what will eventually be passed to
	// the connected minions.
	Src  string `json:"src"`
	url  string
	host string
	path string
	// 'MediaType' tells the connected minions how to embed the desired content
	// using 'Src'.
	MediaType string `json:"mediaType"`
	// Start represents where in the desired content's timeline to begin
	// playing
	Start string `json:"start"`
	// End represents where in the desired content's timeline to stop playing
	End string `json:"end"`
	// Muted represents whether or not the desired content should be muted
	Muted bool `json:"muted"`
}

// SetSrc takes in a string that represents a URL. This function determines if
// the URL is a valid URL, formats imgur links to use .webm instead of .gif(v),
// determines the MediaType that the URL represents, and grabs the videoID from
// YouTube links.
//
// The MediaObj's 'Src' attribute will either be set to the passed URL, the
// formatted imgur URL (if it was an imgur link), or the YouTube video's
// videoID (if it was a YouTube video).
//
// The MediaObj's 'Src' attribute can be retrieved using the MediaObj's
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
	mObj.Src = uri.String()
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

// GetSrc returns the MediaObj's 'Src' attribute (this is what should get passed to
// the connected minions).
func (mObj *MediaObj) GetSrc() string {
	return mObj.Src
}

// GetURL returns the URL that was originally passed to the 'SetSrc()' function.
func (mObj *MediaObj) GetURL() string {
	return mObj.url
}

// setMediaType sets the 'MediaType' attribute of the MediaObj. This tells the
// connected minions what kind of content they should be trying to embed.
func (mObj *MediaObj) setMediaType() {
	// Is the passed URL a YouTube video link?
	if mObj.isYouTube() {
		mObj.MediaType = "youtube"
		return
	}

	// Does the passed URL have a file extension that can be used to determine
	// MediaType?
	matches := reFileExt.FindAllStringSubmatch(mObj.path, -1)
	if len(matches) > 0 {
		fileExt := matches[0][1]

		// Is it an image file?
		for _, ext := range imageFileExts {
			if fileExt == ext {
				mObj.MediaType = "img"
				return
			}
		}

		// Is it an audio file?
		for _, ext := range audioFileExts {
			if fileExt == ext {
				mObj.MediaType = "audio"
				return
			}
		}

		// Is it a video file?
		for _, ext := range videoFileExts {
			if fileExt == ext {
				mObj.MediaType = "video"
				return
			}
		}
	}

	// If it's not a link to a YouTube video, and it's not a link to an image
	// file, audio file, or video file (at least, of those that are supported
	// by Firefox), then make the MediaObj's 'MediaType' be 'web' so the
	// connected minions can just embed the URL as an iframe and hope for the
	// best.
	mObj.MediaType = "web"
	return
}

// GetMediaType returnes the MediaObj's 'MediaType' attribute. The 'MediaType'
// tells the connected minions what kind of content they should be trying to
// embed when using the MediaObj's 'Src' attribute.
func (mObj *MediaObj) GetMediaType() string {
	return mObj.MediaType
}

// IsOfMediaType determines if the MediaObj's MediaType is contained in the
// passed string array
func (mObj *MediaObj) IsOfMediaType(mediaTypes []string) bool {
	mediaType := mObj.GetMediaType()
	for _, mt := range mediaTypes {
		if mediaType == mt {
			return true
		}
	}
	return false
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

// formatImgurURL swaps .gif(v) file extension of MediaObj's Src with .webm
//
// imgur will automatically convert any .gif to .webm, but wrap it as a .gifv.
// WEBM files take far less bandwidth to download than their .gif counterparts,
// and are much easier to render, as well.
//
// Formatting the URL to use .webm (if applicable), allows the URL to be
// recognized as a video file (and thus, will make the 'MediaType' be 'video'),
// so when passed to the connected minions, they can embed it as a video,
// rather than an image.
func (mObj *MediaObj) formatImgurURL() error {
	newURL := reGifV.ReplaceAllString(mObj.Src, ".webm")
	uri, err := parseURL(newURL)
	if err != nil {
		errMsg := "error: " + err.Error() + " (" + mObj.url + ")"
		return errors.New(errMsg)
	}
	mObj.Src = uri.String()
	mObj.path = uri.Path
	return nil
}

// setYouTubeVideoID grabs the YouTube video's videoID from the passed URL, and
// sets it as the MediaObj's 'Src' attribute.
func (mObj *MediaObj) setYouTubeVideoID() {
	mObj.Src = reYTVideoID.FindAllStringSubmatch(mObj.Src, -1)[0][2]
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

// MediaObjList provides a simple way to hold multiple MediaObjs as well as
// references to their intended track, and whether or not the list should loop.
//
// It also allows an easy way to serialize the JSON for the whole list, which
// is what should be passed to the connected minions.
type MediaObjList struct {
	Track     string      `json:"track"`
	Loop      bool        `json:"loop"`
	MediaObjs []*MediaObj `json:"mediaObjs"`
}

// Append appends a pointer to a MediaObj to the end of the 'MediaObjs'
// array.
func (mObjList *MediaObjList) Append(mObj *MediaObj) {
	mObjList.MediaObjs = append(mObjList.MediaObjs, mObj)
}

// Serialize generates and returns the JSON string out of the MediaObjs in the
// 'mediaObj' array. This JSON string is what should be sent to the connected
// minions.
func (mObjList *MediaObjList) Serialize() string {
	serializedJSON, _ := json.Marshal(struct {
		*MediaObjList
		Status string `json:"status"`
	}{
		Status:       "media",
		MediaObjList: mObjList,
	})
	return string(serializedJSON)
}
