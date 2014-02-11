// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"strings"
)

type SayModule struct { }

func (module SayModule) PrivMsg(nick, where, msg string, isAction bool) {
	if strings.HasPrefix(msg, cmd.Nickname+": say ") {
		sendToMinion(msg[6:])
	}
}

func (module SayModule) Init() {
}
