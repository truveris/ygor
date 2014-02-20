// Copyright 2014, Truveris Inc. All Rights Reserved.

package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
)

func PathIsHttp(path string) bool {
	if strings.HasPrefix(path, "http://") {
		return true
	}

	if strings.HasPrefix(path, "https://") {
		return true
	}

	return false
}

func MD5(url string) string {
	h := md5.New()
	io.WriteString(h, url)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func getRemoteSize(url string) int64 {
	resp, err := http.Head(url)
	if err != nil {
		return 0
	}

	return resp.ContentLength
}

// Download the given URL to the given filepath.
func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
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

func mplayer(tune Noise) *exec.Cmd {
	var filepath string
	var cmd *exec.Cmd

	if PathIsHttp(tune.Path) {
		log.Printf("play: path is http")
		size := getRemoteSize(tune.Path)
		if size == 0 {
			log.Printf("play: unable to read HTTP length")
			// say("unable to read HTTP length")
			return nil
		}

		// Too big for local copy, let's stream.
		if size > 512000 {
			filepath = tune.Path
		} else {
			// Check if we already have a copy.
			filepath = "tunes/"+MD5(tune.Path)
			file, err := os.Open(filepath)
			if err != nil {
				log.Printf("play: attempting to cache file...")
				err = downloadFile(tune.Path, filepath)
				if err != nil {
					log.Printf("play: download error:"+
						" %s", err.Error())
					return nil
				}
			}
			file.Close()
		}
	} else {
		filepath = "tunes/" + path.Base(tune.Path)
		if _, err := os.Stat(filepath); err != nil {
			log.Printf("play: stat error: bad filename")
			return nil
		}
	}

	log.Printf("play: path: %s", filepath)

	if cfg.Debug {
		return exec.Command("echo", tune.Duration, filepath)
	}

	if tune.Duration != "" {
		cmd = exec.Command("mplayer", "-really-quiet", "-endpos",
			tune.Duration, filepath)
	} else {
		cmd = exec.Command("mplayer", "-really-quiet", filepath)
	}

	return cmd
}
