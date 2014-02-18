// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"regexp"
	"strings"
)

var (
	reStop = regexp.MustCompile(`^st[aho]+p`)
	reShhh = regexp.MustCompile(`^s+[sh]+`)
)

type ShutUpModule struct{}


func isShutUpRequest(msg string) bool {
	msg = strings.ToLower(msg)
	if reStop.MatchString(msg) {
		return true
	}
	if reShhh.MatchString(msg) {
		return true
	}
	if strings.HasPrefix(msg, "shut up") {
		return true
	}
	return false
}

func shutup(where string) {
	SendToMinion("shutup")
	privMsg(where, "ok...")
}

func (module ShutUpModule) PrivMsg(msg *PrivMsg) {
	if !msg.IsAddressed {
		return
	}

	if isShutUpRequest(msg.Body) {
		shutup(msg.Channel)
		return
	}
}

func (module ShutUpModule) Init() { }
