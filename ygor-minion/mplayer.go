// Copyright 2014-2015, Truveris Inc. All Rights Reserved.

package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tamentis/go-mplayer"
)

const (
	// maxCacheableSize is the maximum file size this minion will keep a
	// local cache of. Anything above this limit will be played directly
	// from the remote location.
	maxCacheableSize = 2 * 1024 * 1024
)

var (
	mplayerInput = make(chan string)
)

// pathIsHTTP checks if the given path appears to be http.
func pathIsHTTP(path string) bool {
	if strings.HasPrefix(path, "http://") {
		return true
	}

	if strings.HasPrefix(path, "https://") {
		return true
	}

	return false
}

// cachedFilename returns the filename of a URL. The current implementation
// returns the MD5 of the URL to create a somewhat unique key.
func cachedFilename(url string) string {
	h := md5.New()
	io.WriteString(h, url)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// Get the Content-Length value for a remote HTTP resource.
func getRemoteSize(url string) (int64, error) {
	resp, err := http.Head(url)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != 200 {
		return 0, errors.New("resp.StatusCode != 200")
	}

	return resp.ContentLength, nil
}

// Download the given URL to the given filepath.
func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("resp.StatusCode != 200")
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func mplayerErrorHandler(err error) {
	log.Printf("play: mplayer exited: %s", err.Error())
	Send("play error: mplayer error: " + err.Error())
}

func mplayerPlayAndWaitWithDuration(filepath string, duration time.Duration) {
	log.Printf("play: play with duration (%s)", duration)
	if cfg.TestMode {
		return
	}
	mplayer.PlayAndWaitWithDuration(filepath, duration)
}

func mplayerPlayAndWait(filepath string) {
	log.Printf("play: play full")
	if cfg.TestMode {
		return
	}
	mplayer.PlayAndWait(filepath)
}

func omxplayer(filepath string) *exec.Cmd {
	log.Printf("play: spawn omxplayer")
	if cfg.TestMode {
		return exec.Command("echo", filepath)
	}

	return exec.Command("omxplayer", filepath)
}

// If any player process is instantiated, this function will return the live
// cmd process. In any other case (error or if we are using the mplayer
// module), it returns nil.
func player(tune Noise) *exec.Cmd {
	var filepath string

	// Videos are stored locally until we have a youtube mirroring service.
	if strings.HasPrefix(filepath, "video") {
		return omxplayer(filepath)
	}

	// Everything outside of videos should be accessible via a URL and is
	// cached locally prior to being played.  This is mostly because
	// mplayer is not able to play from https.
	if !pathIsHTTP(tune.Path) {
		Send("play error invalid URL")
		return nil
	}

	// Check if we have a local copy, if not make one.
	filepath = "tunes/" + cachedFilename(tune.Path)
	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("play: downloading file to local cache")
		err = downloadFile(tune.Path, filepath)
		if err != nil {
			Send("play caching error")
			log.Printf("play: download error:"+
				" %s, reverting to streaming", err.Error())
			filepath = tune.Path
		}
	}
	file.Close()

	log.Printf("play: path=%s duration=%d", filepath, tune.Duration)
	if tune.Duration != 0 {
		mplayerPlayAndWaitWithDuration(filepath, tune.Duration)
	} else {
		mplayerPlayAndWait(filepath)
	}

	return nil
}
