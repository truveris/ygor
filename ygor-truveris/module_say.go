// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"strings"
)

type SayModule struct { }

func (module SayModule) PrivMsg(msg *PrivMsg) {
	// Turn that shit into a command.
	if msg.IsAddressed && strings.HasPrefix(msg.Body, "say ") {
		SendToMinion(msg.Body)
	}
}

func (module SayModule) Init() {
}
