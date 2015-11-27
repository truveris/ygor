// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"errors"
	"regexp"
)

// ParseArgList validates command usage, and returns a map containing the
// arguments for the media item
func parseArgList(msgArgs []string) (map[string]string, error) {
	// Validate that it has an appropriate number of arguments.
	numOfArgs := len(msgArgs)
	if numOfArgs < 1 || numOfArgs > 2 {
		err := errors.New("invalid number of arguments")
		return nil, err
	}
	// Start making the map that represents this media item.
	m := make(map[string]string)
	// The URL should be the first argument. If the first argument isn't a URL,
	// it will be determined and handled elsewhere.
	m["url"] = msgArgs[0]
	// Get the duration passed for this media item (if it was passed).
	duration, err := getDuration(msgArgs)
	if err != nil {
		return nil, err
	}
	// Insert the duration into the map.
	m["duration"] = duration
	// Return the completed map.
	return m, nil
}

// getDuration grabs the duration that the media item is to be played, if it is
// passed. If a duration isn't passed, an empty string is returned.
func getDuration(args []string) (string, error) {
	reTime := regexp.MustCompile(`^(([0-9]*\.)?[0-9]+)`)
	// Set the default string for duration to an empty string to simplify
	// things.
	duration := ""
	if len(args) == 2 {
		duration = reTime.FindStringSubmatch(args[1])[1]
	}

	return duration, nil
}
