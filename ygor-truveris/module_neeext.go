// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// Randomly scream "NEXT!!" for no reason, like the original Ygor.

package main

import (
	"time"
	"math/rand"
)

type NeeextModule struct {
	LastSpeaker string
	LastChannel string
}

func (module NeeextModule) PrivMsg(nick, where, msg string, isAction bool) {
	if nick != cmd.Nickname && nick != module.LastSpeaker {
		module.LastSpeaker = nick
		module.LastChannel = where
	}
}

func (module NeeextModule) Gibberish() {
	privAction(module.LastChannel, "stares at "+module.LastSpeaker+"...")
	time.Sleep(2 * time.Second)
	if rand.Float64() > 0.5 {
		privMsg(module.LastChannel, "NEXT!!")
	} else {
		privMsg(module.LastChannel, "CHICKEN SANDWICH!!")
	}
}

func (module NeeextModule) Ticker() {
	ticker := time.Tick(1 * time.Hour)
	chances := 1.0 / 1000.0

	for _ = range ticker {
		if rand.Float64() < chances {
			module.Gibberish()
		}
	}
}

func (module NeeextModule) Init() {
	go module.Ticker()
}
