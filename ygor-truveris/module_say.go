// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"strings"
)

type SayModule struct { }

func (module SayModule) PrivMsg(msg *PrivMsg) {
	if strings.HasPrefix(msg.Body, cmd.Nickname+": say ") {
		SendToMinion(msg.Body[6:])
	}
}

func (module SayModule) Init() {
}
