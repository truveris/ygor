// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"errors"
	"regexp"
	"strings"
)

// validates command usage, and returns a map array containing the arguments
// for each media item in the list
func parseArgList(argArr []string) ([]map[string]string, error) {
	// make it a single string
	mediaListStr := strings.Join(argArr, " ")
	// media list should be deliminated by commas
	mediaListArr := regexp.MustCompile(`,\s*`).Split(mediaListStr, -1)
	mediaList := []map[string]string{}
	for _, media := range mediaListArr {
		mediaArgs := strings.Split(media, " ")
		if len(mediaArgs) < 1 || len(mediaArgs) > 3 {
			err := errors.New("invalid number of arguments")
			return mediaList, err
		}
		m := make(map[string]string)
		m["url"] = mediaArgs[0]
		start, end, err := getBounds(mediaArgs)
		if err != nil {
			return mediaList, err
		}
		m["start"] = start
		m["end"] = end
		mediaList = append(mediaList, m)
	}
	return mediaList, nil
}

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
