// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"path"
	"regexp"
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
	gfycatHostNames = []string{
		"gfycat.com",
		"www.gfycat.com",
		"zippy.gfycat.com",
		"fat.gfycat.com",
		"giant.gfycat.com",
		"center.gfycat.com",
		"center2.gfycat.com",
		"centre.gfycat.com",
		"test.gfycat.com",
		"upload.gfycat.com",
	}
	youtubeHostNames = []string{
		"www.youtube.com",
		"www.youtu.be",
		"youtube.com",
		"youtu.be",
	}
	vimeoHostNames = []string{
		"vimeo.com",
		"www.vimeo.com",
		"player.vimeo.com",
		"www.player.vimeo.com",
	}
	soundcloudHostNames = []string{
		"soundcloud.com",
		"www.soundcloud.com",
		"api.soundcloud.com",
		"www.api.soundcloud.com",
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

	// ygor should fallback to checking the file extensions for potential
	// matches if the content-type doesn't appear to be supported. The server
	// may simply be providing the wrong content-type in the header.
	supportedFormatsAndExtensions = map[string][]string{
		"img": {
			".apng",
			".bmp",
			".dib",
			".gif",
			".jfi",
			".jfif",
			".jif",
			".jpe",
			".jpeg",
			".jpg",
			".png",
			".webp",
		},
		"audio": {
			".mp3",
			".wav",
			".wave",
		},
		"video": {
			".m4a",
			".m4b",
			".m4p",
			".m4r",
			".m4v",
			".mp4",
			".oga",
			".ogg",
			".ogm",
			".ogv",
			".ogx",
			".opus",
			".spx",
			".webm",
		},
	}

	reYTVideoID = regexp.MustCompile(
		`^.*(youtu.be\/|v\/|u\/\w\/|embed\/|watch\?v=|\&v=)([^#\&\?]*).*`)

	// Reserved private IP address ranges
	privIPRanges = map[string]map[string][]byte{
		"RFC 1968 A": {
			"low": net.ParseIP("10.0.0.0"),
			"high": net.ParseIP("10.255.255.255"),
		},
		"RFC 1968 B": {
			"low": net.ParseIP("172.16.0.0"),
			"high": net.ParseIP("172.31.255.255"),
		},
		"RFC 1968 C": {
			"low": net.ParseIP("192.168.0.0"),
			"high": net.ParseIP("192.168.255.255"),
		},
		"RFC 6958": {
			"low": net.ParseIP("100.64.0.0"),
			"high": net.ParseIP("100.127.255.255"),
		},
	}

)

// Media represents the relevant data that will eventually be passed to
// the connected minions. It is used to generate the information that connected
// minions would use to properly embed the desired content.
//
// It also provides several functions that can be used to more easily work with
// the data, so that command modules aren't filled with a lot of excessive
// code.
type Media struct {
	// 'Src' is formatted over time and is what will eventually be passed to
	// the connected minions.
	Src  string `json:"src"`
	url  string
	host string
	// 'Format' tells the connected minions how to embed the desired content
	// using 'Src'.
	Format    string `json:"format"`
	mediaType string
	// End represents where in the desired content's timeline to stop playing.
	End string `json:"end"`
	// Muted represents whether or not the desired content should be muted.
	Muted bool `json:"muted"`
	Loop  bool `json:"loop"`
	track string
	// acceptableFormats is a list of acceptable media types, it which
	// will be checked against during SetSrc. If the determined media type
	// is not acceptable, the url will be rejected.
	acceptableFormats []string
	// srv provides easy access to the Server, just in case.
	srv *Server
}

// checkFormatIsAcceptable checks to make sure that the determined media
// type is acceptable. If the Media's acceptableFormats attribute is not
// set, it is assumed that the media type is acceptable.
func (media *Media) checkFormatIsAcceptable() error {
	if len(media.acceptableFormats) == 0 {
		// if acceptableFormats is not set, all media types are acceptable
		return nil
	}

	for _, acceptableFormat := range media.acceptableFormats {
		if media.Format == acceptableFormat {
			// The determined media type is acceptable.
			return nil
		}
	}

	// If it made it here, the determined media type must not be acceptable.
	errMsg := "error: content-type (" + media.mediaType + ") not supported " +
		"by this command"
	return errors.New(errMsg)
}

// SetSrc takes in a string that represents a URL. This function determines if
// the URL is a valid URL, formats imgur links to use .webm instead of .gif(v),
// and determines the Format that the URL represents.
//
// The Media's 'Src' attribute will either be set to the passed URL, or the
// formatted imgur URL (if it was an imgur link).
//
// The Media's 'Src' attribute can be retrieved using the Media's
// 'GetSrc()' function.
//
// The URL that was originally passed, is saved as the Media's 'url'
// attribute, and can be retrieved with the Media's 'GetURL()' function.
func (media *Media) SetSrc(link string) error {
	uri, linkErr := url.ParseRequestURI(link)
	if linkErr != nil {
		errorMsg := "error: not a valid URL"
		return errors.New(errorMsg)
	}
	// Strip any query or fragment attached to the URL
	media.Src = uri.String()
	media.url = link
	media.host = uri.Host

	var header map[string][]string

	if !hostIsPrivateIP(media.host) {
		// Check that the URL returns a status code of 200.
		res, err := http.Head(media.Src)
		if err != nil {
			errMsg := "error: " + err.Error()
			return errors.New(errMsg)
		}
		statusCode := strconv.Itoa(res.StatusCode)
		if statusCode != "200" {
			errMsg := "error: response status code is " + statusCode
			return errors.New(errMsg)
		}
		header = res.Header
	} else {
		header = map[string][]string{
			"Content-Type": {
				"unknown",
			},
		}
	}

	headErr := media.setFormat(header)
	if headErr != nil {
		return headErr
	}

	// If it's an imgur link, and the content-type contains "image/gif", modify
	// the Media so minions embed the far more efficient webm version.
	if media.isImgur() {
		isGIF := strings.Contains(strings.ToLower(media.mediaType), "image/gif")
		hasGIFVExt := media.GetExt() == ".gifv"
		if (isGIF || hasGIFVExt) && media.hasWebm() {
			media.replaceSrcExt(".webm")
			media.Format = "video"
			media.mediaType = "video/webm"
		}
	}

	// If it's a Gfycat link, and the content-type isn't "video/webm", attempt
	// to find a webm version of this link using the Gfycat API.
	if media.isGfycat() {
		isWEBM := strings.Contains(strings.ToLower(media.mediaType), "video/webm")
		if !isWEBM {
			media.resolveGfycatURL()
		}
	}

	merr := media.checkFormatIsAcceptable()
	if merr != nil {
		return merr
	}

	return nil
}

// GetSrc returns the Media's 'Src' attribute (this is what should get
// passed to the connected minions).
func (media *Media) GetSrc() string {
	return media.Src
}

// GetURL returns the URL that was originally passed to the 'SetSrc()'
// function.
func (media *Media) GetURL() string {
	return media.url
}

// hostIsPrivateIP checks if the passed host is an IP address from a private IP
// address range. If so, it returns true. Otherwise, it returns false.
// 
// This function is useful for determining if ygor can make a request at all to
// the passed URL.
func hostIsPrivateIP(host string) bool {
	// Strip the port if there is one.
	portIndex := strings.Index(host, ":")
	if portIndex != -1 {
		// There is a port.
		host = host[:portIndex]
	}
	ip := net.ParseIP(host)
	if ip.To4() == nil {
		// This is not an IPv4 address.
		return false
	}
	for _, IPRange := range privIPRanges {
		if bytes.Compare(ip, IPRange["low"]) >= 0 && bytes.Compare(ip, IPRange["high"]) <= 0 {
			// This is an IP address that is in a private range.
			return true
		}
	}
	return false
}

// setFormat sets the 'Format' attribute of the Media. This tells the
// connected minions what kind of content they should be trying to embed.
func (media *Media) setFormat(header map[string][]string) error {
	// If it's a YouTube link, check if there's a video ID we can grab.
	if media.isYouTube() {
		match := reYTVideoID.FindAllStringSubmatch(media.Src, -1)
		if len(match) > 0 {
			media.Src = match[0][2]
			media.Format = "youtube"
			media.mediaType = "youtube"
			return nil
		}
	}

	// If it's a Vimeo link, check if there's a video ID we can grab.
	if media.isVimeo() {
		// Vimeo video IDs are the last element in the URL (represented as an
		// integer between 6 and 11 digits long) before the query string and/or
		// fragment. media.Src has the query string and fragment stripped off,
		// so if this is a link to a Vimeo video, potentialVideoID should be an
		// integer between 6 and 11 digits long.
		potentialVideoID := path.Base(media.Src)
		// Check to see if it is between 6 and 11 characters long.
		if 6 <= len(potentialVideoID) && len(potentialVideoID) <= 11 {
			// Check to make sure it is a number.
			if _, err := strconv.Atoi(potentialVideoID); err == nil {
				// It is a number
				media.Src = potentialVideoID
				media.Format = "vimeo"
				media.mediaType = "vimeo"
				return nil
			}
		}
	}

	// If it's a SoundCloud URL, attempt to resolve it and get the link to
	// embed the song.
	if media.isSoundCloud() {
		resolveErr := media.resolveSoundCloudURL()
		if resolveErr != nil {
			return resolveErr
		}
		if media.GetFormat() == "soundcloud" {
			// The link was to a SoundCloud track.
			return nil
		}
		// If it isn't a link to a SoundCloud track, continue on and handle the
		// URL like any other.
	}

	// Is the media type in the contentType an image|audio|video type that
	// Chromium supports?
	if contentType, ok := header["Content-Type"]; ok {
		// Check for standard, supported media types.
		for format, formatMediaTypes := range supportedFormatsAndTypes {
			for _, mediaType := range formatMediaTypes {
				for _, cType := range contentType {
					if strings.Contains(cType, mediaType) {
						media.Format = format
						media.mediaType = mediaType
						return nil
					}
				}
			}
		}

		// Fallback to known supported file extensions if content-type isn't
		// recognized as supported.
		ext := media.GetExt()
		for format, formatExtensions := range supportedFormatsAndExtensions {
			for _, extension := range formatExtensions {
				if extension == ext {
					media.Format = format
					media.mediaType = ext
					return nil
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

// GetFormat returnes the Media's 'Format' attribute. The 'Format'
// tells the connected minions what kind of content they should be trying to
// embed when using the Media's 'Src' attribute.
func (media *Media) GetFormat() string {
	return media.Format
}

// IsOfFormat determines if the Media's Format is contained in the
// passed string array.
func (media *Media) IsOfFormat(formats []string) bool {
	format := media.GetFormat()
	for _, mt := range formats {
		if format == mt {
			return true
		}
	}
	return false
}

// GetExt is a convenience function to get the extension of theMedia's
// current Src.
func (media *Media) GetExt() string {
	cleanSrc := media.Src
	// Strip the fragment if there is one.
	fragmentIndex := strings.Index(cleanSrc, "#")
	if fragmentIndex != -1 {
		// There is a fragment.
		cleanSrc = cleanSrc[:fragmentIndex]
	}
	// Strip the query string if there is one.
	queryIndex := strings.Index(cleanSrc, "?")
	if queryIndex != -1 {
		// There is a query string.
		cleanSrc = cleanSrc[:queryIndex]
	}
	return strings.ToLower(path.Ext(cleanSrc))
}

// isImgur attempts to determine if the desired content is hosted on imgur.
func (media *Media) isImgur() bool {
	for _, d := range imgurHostNames {
		if media.host == d {
			return true
		}
	}
	return false
}

// isGfycat attempts to determine if the desired content is hosted on gfycat.
func (media *Media) isGfycat() bool {
	for _, d := range gfycatHostNames {
		if media.host == d {
			return true
		}
	}
	return false
}

// isYouTube attempts to determine if the desired content is a video hosted on
// YouTube
func (media *Media) isYouTube() bool {
	for _, d := range youtubeHostNames {
		if media.host == d {
			return true
		}
	}
	return false
}

// isVimeo attempts to determine if the desired content is a video hosted on
// Vimeo
func (media *Media) isVimeo() bool {
	for _, d := range vimeoHostNames {
		if media.host == d {
			return true
		}
	}
	return false
}

// isSoundCloud attempts to determine if the desired content is a song hosted
// on SoundCloud.
func (media *Media) isSoundCloud() bool {
	for _, d := range soundcloudHostNames {
		if media.host == d {
			return true
		}
	}
	return false
}

// hasWebm checks if there is a webm version of the provided URL. It returns
// true if there is, and false if there isn't.
func (media *Media) hasWebm() bool {
	webmURL := media.Src[0:len(media.Src)-len(media.GetExt())] + ".webm"
	// Check that the URL returns a status code of 200.
	res, err := http.Head(webmURL)
	if err != nil {
		// There likely isn't a webm version.
		return false
	}
	statusCode := strconv.Itoa(res.StatusCode)
	if statusCode != "200" {
		// There likely isn't a webm version.
		return false
	}
	// A webm version has been found.
	return true
}

// resolveSoundCloudURL attempts to find the track URI of the SoundCloud link
// that was provided using SoundCloud's API along with the SoundCloudClientID
// saved in the config file. If it finds a track URI, this is saved as the
// Media's Src and its Format is set to "soundcloud". If there isn't a track
// URI, this function does not return an error. Instead, it relies on the
// function calling it to determne whether or not the Media's Format was set
// to "soundcloud".
func (media *Media) resolveSoundCloudURL() error {
	if media.srv.Config.SoundCloudClientID == "" {
		errMsg := "error: SoundCloudClientID is not configured"
		return errors.New(errMsg)
	}
	resolveURL := "http://api.soundcloud.com/resolve?url=" + media.Src +
		"&client_id=" + media.srv.Config.SoundCloudClientID
	// Make the request.
	res, err := http.Get(resolveURL)
	if err != nil {
		errMsg := "error: " + err.Error()
		return errors.New(errMsg)
	}
	rBody, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		errMsg := "error: malformed SoundCloud API response " +
			"(could not read all)"
		return errors.New(errMsg)
	}
	var dat map[string]interface{}
	if err := json.Unmarshal(rBody, &dat); err != nil {
		errMsg := "error: malformed SoundCloud API response " +
			"(could not parse)"
		return errors.New(errMsg)
	}
	if kind, hasKind := dat["kind"]; hasKind {
		if kind == "track" {
			if trackURI, hasTrackURI := dat["uri"]; hasTrackURI {
				media.Src = trackURI.(string)
				media.Format = "soundcloud"
				media.mediaType = "soundcloud"
				return nil
			}
			// If the API response says the kind is 'track', but there
			// is no uri, return an error.
			errMsg := "error: malformed SoundCloud API response " +
				"(no track uri)"
			return errors.New(errMsg)
		}
	}
	// If it hasn't returned by now, it isn't a link to a SoundCloud track.
	return nil
}

// resolveGfycatURL attempts to find get the URL of the webm version of the
// provided Gfycat URL. If there is no webm URL, it falls back to the mp4 URL.
// If there is no mp4 URL, it then falls back to the gif URL. If the Gfycat API
// doesn't return any JSON to be parsed, then it isn't a link to a Gfycat
// image/video, so nothing should be changed in the Media.
func (media *Media) resolveGfycatURL() error {
	gfyName := path.Base(media.Src)
	resolveURL := "http://gfycat.com/cajax/get/" + gfyName
	// Make the request.
	res, err := http.Get(resolveURL)
	if err != nil {
		errMsg := "error: " + err.Error()
		return errors.New(errMsg)
	}
	rBody, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		errMsg := "error: malformed Gfycat API response " +
			"(could not read all)"
		return errors.New(errMsg)
	}
	var dat map[string]map[string]interface{}
	if err := json.Unmarshal(rBody, &dat); err != nil {
		errMsg := "error: malformed Gfycat API response " +
			"(could not parse)"
		return errors.New(errMsg)
	}
	if gfyItem, hasGfyItem := dat["gfyItem"]; hasGfyItem {
		if webmURL, hasWebmURL := gfyItem["webmUrl"]; hasWebmURL {
			media.Src = webmURL.(string)
			media.Format = "video"
			media.mediaType = "video/webm"
			return nil
		} else if mp4URL, hasMp4URL := gfyItem["mp4Url"]; hasMp4URL {
			// If, for some reason, there isn't a webm URL, fallback to the mp4
			// URL.
			media.Src = mp4URL.(string)
			media.Format = "video"
			media.mediaType = "video/mp4"
			return nil
		} else if gifURL, hasGifURL := gfyItem["gifUrl"]; hasGifURL {
			// If, for some reason, there isn't an mp4 URL either, fallback to
			// the gif URL.
			media.Src = gifURL.(string)
			media.Format = "image"
			media.mediaType = "image/gif"
			return nil
		}
		// If the API response says it has a 'gfyItem', but doesn't provide any
		// URLs related to it, return an error.
		errMsg := "error: malformed Gfycat API response " +
			"(no content URLs provided)"
		return errors.New(errMsg)
	}
	// If it hasn't returned by now, it isn't a link to a Gfycat track.
	return nil
}

// replaceSrcExt is a convenience function to replace the extension of the
// Media's current Src.
func (media *Media) replaceSrcExt(newExt string) {
	media.Src = media.Src[0:len(media.Src)-len(media.GetExt())] + newExt
}

// Serialize generates and returns the JSON string out of the Media. This
// JSON string is what should be sent to the connected minions.
func (media *Media) Serialize() string {
	serializedJSON, _ := json.Marshal(struct {
		Media  *Media `json:"media"`
		Status string `json:"status"`
		Track  string `json:"track"`
	}{
		Status: "media",
		Track:  media.track,
		Media:  media,
	})
	return string(serializedJSON)
}

// NewMedia is a convenience function meant to clean up the code of modules.
// It builds the Media.
func NewMedia(srv *Server, mediaItem map[string]string, track string, muted bool, loop bool, acceptableFormats []string) (*Media, error) {
	// Parse the mediaItem map into a Media.
	media := new(Media)
	media.srv = srv
	media.End = mediaItem["end"]
	media.Muted = muted
	media.Loop = loop
	media.track = track
	media.acceptableFormats = acceptableFormats

	setSrcErr := media.SetSrc(mediaItem["url"])
	if setSrcErr != nil {
		return nil, setSrcErr
	}

	return media, nil
}
