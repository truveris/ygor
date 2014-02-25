// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
)

type RepeatModule struct{}

func (module RepeatModule) PrivMsg(nick, where, msg string, isAction bool) {
	if isAction {
		msg = "/ME " + msg
	}
	newMsg := fmt.Sprintf("nick:%s channel:%s msg:%s", nick, where, msg)
	privMsg(where, newMsg)
}

func (module RepeatModule) Init() {
}
