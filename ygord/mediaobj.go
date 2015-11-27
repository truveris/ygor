// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

var (
	// These are the known domains to check for, where special formatting of
	// the passed URL is required so connected minions can most effectively
	// embed and manipulate the desired content.
	imgurHostNames = []string{
		"i.imgur.com",
		"www.imgur.com",
		"imgur.com",
	}

	supportedFormatsAndTypes = map[string][]string{
		"img": {
			"image/bmp",
			"image/cis-cod",
			"image/gif",
			"image/ief",
			"image/jpeg",
			"image/webp",
			"image/pict",
			"image/pipeg",
			"image/png",
			"image/svg+xml",
			"image/tiff",
			"image/vnd.microsoft.icon",
			"image/x-cmu-raster",
			"image/x-cmx",
			"image/x-icon",
			"image/x-portable-anymap",
			"image/x-portable-bitmap",
			"image/x-portable-graymap",
			"image/x-portable-pixmap",
			"image/x-rgb",
			"image/x-xbitmap",
			"image/x-xpixmap",
			"image/x-xwindowdump",
		},
		"audio": {
			"audio/aac",
			"audio/aiff",
			"audio/amr",
			"audio/basic",
			"audio/midi",
			"audio/mp3",
			"audio/mp4",
			"audio/mpeg",
			"audio/mpeg3",
			"audio/ogg",
			"audio/vorbis",
			"audio/wav",
			"audio/webm",
			"audio/x-m4a",
			"audio/x-ms-wma",
			"audio/vnd.rn-realaudio",
			"audio/vnd.wave",
		},
		"video": {
			"video/avi",
			"video/divx",
			"video/flc",
			"video/mp4",
			"video/mpeg",
			"video/ogg",
			"video/quicktime",
			"video/sd-video",
			"video/webm",
			"video/x-dv",
			"video/x-m4v",
			"video/x-mpeg",
			"video/x-ms-asf",
			"video/x-ms-wmv",
		},
		"web": {
			"text/",
		},
	}
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
	// 'Format' tells the connected minions how to embed the desired content
	// using 'Src'.
	Format      string `json:"format"`
	mediaType   string
	// End represents where in the desired content's timeline to stop playing.
	End string `json:"end"`
	// Muted represents whether or not the desired content should be muted.
	Muted             bool `json:"muted"`
	Loop              bool `json:"loop"`
	track             string
	acceptableFormats []string
}

// SetAcceptableFormats takes in a string array of acceptable media types,
// which will be checked against during SetSrc. If the determined media type is
// not acceptable, the url will be rejected.
func (mObj *MediaObj) SetAcceptableFormats(formats []string) {
	mObj.acceptableFormats = formats
}

// checkFormatIsAcceptable checks to make sure that the determined media
// type is acceptable. If the MediaObj's acceptableFormats attribute is not
// set, it is assumed that the media type is acceptable.
func (mObj *MediaObj) checkFormatIsAcceptable() error {
	if len(mObj.acceptableFormats) == 0 {
		// if acceptableFormats is not set, all media types are acceptable
		return nil
	}

	for _, acceptableFormat := range mObj.acceptableFormats {
		if mObj.Format == acceptableFormat {
			// The determined media type is acceptable.
			return nil
		}
	}

	// If it made it here, the determined media type must not be acceptable.
	errMsg := "error: content-type (" + mObj.mediaType + ") not supported " +
		"by this command"
	return errors.New(errMsg)
}

// SetSrc takes in a string that represents a URL. This function determines if
// the URL is a valid URL, formats imgur links to use .webm instead of .gif(v),
// and determines the Format that the URL represents.
//
// The MediaObj's 'Src' attribute will either be set to the passed URL, or the
// formatted imgur URL (if it was an imgur link).
//
// The MediaObj's 'Src' attribute can be retrieved using the MediaObj's
// 'GetSrc()' function.
//
// The URL that was originally passed, is saved as the MediaObj's 'url'
// attribute, and can be retrieved with the MediaObj's 'GetURL()' function.
func (mObj *MediaObj) SetSrc(link string) error {
	uri, linkErr := url.ParseRequestURI(link)
	if linkErr != nil {
		errorMsg := "error: not a valid URL"
		return errors.New(errorMsg)
	}
	// Strip any query or fragment attached to the URL
	mObj.Src = uri.String()
	mObj.url = link
	mObj.host = uri.Host

	// Check that the URL returns a status code of 200.
	res, err := http.Head(mObj.Src)
	if err != nil {
		errMsg := "error: " + err.Error()
		return errors.New(errMsg)
	}
	statusCode := strconv.Itoa(res.StatusCode)
	if statusCode != "200" {
		errMsg := "error: response status code is " + statusCode
		return errors.New(errMsg)
	}

	headErr := mObj.setFormat(res.Header)
	if headErr != nil {
		return headErr
	}

	// If it's an imgur link, and the content-type contains "image/gif", modify
	// the MediaObj so minions embed the far more efficient webm version.
	if mObj.isImgur() {
		isGIF := strings.Contains(strings.ToLower(mObj.mediaType), "image/gif")
		hasGIFVExt := mObj.GetExt() == ".gifv"
		if isGIF || hasGIFVExt {
			mObj.replaceSrcExt(".webm")
			mObj.Format = "video"
			mObj.mediaType = "video/webm"
		}
	}

	merr := mObj.checkFormatIsAcceptable()
	if merr != nil {
		return merr
	}

	return nil
}

// GetSrc returns the MediaObj's 'Src' attribute (this is what should get
// passed to the connected minions).
func (mObj *MediaObj) GetSrc() string {
	return mObj.Src
}

// GetURL returns the URL that was originally passed to the 'SetSrc()'
// function.
func (mObj *MediaObj) GetURL() string {
	return mObj.url
}

// setFormat sets the 'Format' attribute of the MediaObj. This tells the
// connected minions what kind of content they should be trying to embed.
func (mObj *MediaObj) setFormat(header map[string][]string) error {
	// Is the media type in the contentType an image|audio|video type that
	// Chromium supports?
	if contentType, ok := header["Content-Type"]; ok {
		// Check for standard, supported media types.
		for format, formatMediaTypes := range supportedFormatsAndTypes {
			for _, mediaType := range formatMediaTypes {
				for _, cType := range contentType {
					if strings.Contains(cType, mediaType) {
						mObj.Format = format
						mObj.mediaType = mediaType
						return nil
					}
				}
			}
		}

		// If the media type isn't supported, return an error.
		errMsg := "error: unsupported content-type " +
			"(" + strings.Join(contentType, ", ") + ")"
		return errors.New(errMsg)
	}

	// It will only get here if it didn't have a content-type in the header.
	errMsg := "error: no content-type found"
	return errors.New(errMsg)
}

// GetFormat returnes the MediaObj's 'Format' attribute. The 'Format'
// tells the connected minions what kind of content they should be trying to
// embed when using the MediaObj's 'Src' attribute.
func (mObj *MediaObj) GetFormat() string {
	return mObj.Format
}

// IsOfFormat determines if the MediaObj's Format is contained in the
// passed string array.
func (mObj *MediaObj) IsOfFormat(formats []string) bool {
	format := mObj.GetFormat()
	for _, mt := range formats {
		if format == mt {
			return true
		}
	}
	return false
}

// GetExt is a convenience function to get the extension of theMediaObj's
// current Src.
func (mObj *MediaObj) GetExt() string {
	return strings.ToLower(path.Ext(mObj.Src))
}

// isImgur attempts to determine if the desired content is hosted on imgur.
func (mObj *MediaObj) isImgur() bool {
	for _, d := range imgurHostNames {
		if mObj.host == d {
			return true
		}
	}
	return false
}

// replaceSrcExt is a convenience function to replace the extension of the
// MediaObj's current Src.
func (mObj *MediaObj) replaceSrcExt(newExt string) {
	mObj.Src = mObj.Src[0:len(mObj.Src)-len(mObj.GetExt())] + newExt
}

// Serialize generates and returns the JSON string out of the MediaObj. This
// JSON string is what should be sent to the connected minions.
func (mObj *MediaObj) Serialize() string {
	serializedJSON, _ := json.Marshal(struct {
		MediaObj *MediaObj `json:"mediaObj"`
		Status   string    `json:"status"`
		Track    string    `json:"track"`
	}{
		Status:   "media",
		Track:    mObj.track,
		MediaObj: mObj,
	})
	return string(serializedJSON)
}

// NewMediaObj is a convenience function meant to clean up the code of modules.
// It builds the MediaObj.
func NewMediaObj(mediaItem map[string]string, track string, muted bool, loop bool, acceptableFormats []string) (*MediaObj, error) {
	// Parse the mediaItem map into a MediaObj.
	mObj := new(MediaObj)
	mObj.End = mediaItem["end"]
	mObj.Muted = muted
	mObj.Loop = loop
	mObj.track = track
	mObj.SetAcceptableFormats(acceptableFormats)

	setSrcErr := mObj.SetSrc(mediaItem["url"])
	if setSrcErr != nil {
		return nil, setSrcErr
	}

	return mObj, nil
}
