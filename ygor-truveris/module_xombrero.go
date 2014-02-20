// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"strings"
)

type XombreroModule struct{}

func (module XombreroModule) PrivMsg(msg *PrivMsg) { }

func Xombrero(where string, params []string) {
	if len(params) == 0 {
		privMsg(where, "usage: xombrero [command [param ...]]")
		return
	}

	SendToMinion("xombrero "+strings.Join(params, " "))
	privMsg(where, "sure")
}


func (module XombreroModule) Init() {
	RegisterCommand(NewCommandFromFunction("xombrero", Xombrero))
}
