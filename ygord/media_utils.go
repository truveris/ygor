// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"errors"
	"regexp"
	"strings"
)

// ParseArgList validates command usage, and returns a map array containing the
// arguments for each media item in the list.
func parseArgList(argArr []string) ([]map[string]string, error) {
	// Combine the arguments into a single string
	mediaListStr := strings.Join(argArr, " ")
	// Now split at the commas to separate each media item so they can each be
	// validated individually.
	mediaListArr := regexp.MustCompile(`,\s*`).Split(mediaListStr, -1)
	// Make the map array that will house the formatted media items.
	mediaList := []map[string]string{}
	for _, media := range mediaListArr {
		// Split the individual media item at the spaces to seperate the
		// arguments.
		mediaArgs := strings.Split(media, " ")
		// Validate that it has an appropriate number of arguments.
		numOfArgs := len(mediaArgs)
		if numOfArgs != 1 && numOfArgs != 3 && numOfArgs != 5 {
			err := errors.New("invalid number of arguments")
			return mediaList, err
		}
		// Start making the map that represents this media item.
		m := make(map[string]string)
		// Insert the URL of this media item into the map that represents it.
		m["url"] = mediaArgs[0]
		// Get the start and end bounds passed for this media item (if they
		// were passed).
		start, end, err := getBounds(mediaArgs)
		if err != nil {
			return mediaList, err
		}
		// Insert the start and end bounds into the map that represents this
		// media item.
		m["start"] = start
		m["end"] = end
		// Append the now completely formed map that represents this media item
		// into the map array that repesents all the media items passed by this
		// command.
		mediaList = append(mediaList, m)
	}
	// Return the completed map array that represents all the media items
	// passed by this command.
	return mediaList, nil
}

// getBounds grabs starting and ending time frame bounds if either is passed.
// If a bound isn't passed, an empty string is returned to represent that
// bound.
func getBounds(args []string) (string, string, error) {
	sBound := ""
	eBound := ""
	if len(args) == 3 || len(args) == 5 {
		// Grab and validate the first option.
		if args[1] == "-s" {
			sBound = args[2]
		} else if args[1] == "-e" {
			eBound = args[2]
		} else {
			return "", "", errors.New("invalid argument")
		}
		if len(args) == 5 {
			// Grab and validate the second option.
			if args[3] == "-s" && args[1] == "-e" {
				sBound = args[4]
			} else if args[3] == "-e" && args[1] == "-s" {
				eBound = args[4]
			} else {
				return "", "", errors.New("invalid argument")
			}
		}
	}

	return sBound, eBound, nil
}
