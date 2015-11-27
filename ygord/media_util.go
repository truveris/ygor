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
	// Get the end passed for this media item (if it was passed).
	end, err := getEnd(msgArgs)
	if err != nil {
		return nil, err
	}
	// Insert the end into the map.
	m["end"] = end
	// Return the completed map.
	return m, nil
}

// getEnd grabs the end that the media item is to be played, if it is
// passed. If a end isn't passed, an empty string is returned.
func getEnd(args []string) (string, error) {
	reTime := regexp.MustCompile(`^(([0-9]*\.)?[0-9]+)`)
	// Set the default string for end to an empty string to simplify
	// things.
	end := ""
	if len(args) == 2 {
		end = reTime.FindStringSubmatch(args[1])[1]
	}

	return end, nil
}
