// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"regexp"
)

var (
	reAddressed = regexp.MustCompile(`^(\w+)[:,.]+\s+(.*)`)
)

// Returns the PRIVMSG without the nickname prefix if any, if the message was
// not addressed to this bot, it returns an empty string.
func AddressedToMe(msg string) string {
	tokens := reAddressed.FindStringSubmatch(msg)
	if tokens == nil {
		return ""
	}

	if tokens[1] == cmd.Nickname {
		return tokens[2]
	}

	return ""
}
