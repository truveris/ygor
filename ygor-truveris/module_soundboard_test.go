// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"
)

func TestReSoundBoard(t *testing.T) {
	var tokens = make([]string, 0)

	tokens = reSoundBoard.FindStringSubmatch("anything")
	if tokens == nil {
		t.Fatal("couldn't match 'anything'")
	}

	tokens = reSoundBoard.FindStringSubmatch("wagner")
	if tokens[1] != "wagner"{
		t.Fatal("couldn't match 'wagner' by itself:", tokens)
	}

	tokens = reSoundBoard.FindStringSubmatch("wagner 15")
	if tokens[1] != "wagner"{
		t.Fatal("couldn't match 'wagner 15' by itself:", tokens)
	}

	tokens = reSoundBoard.FindStringSubmatch("play wagner")
	if tokens[1] != "wagner"{
		t.Fatal("couldn't match 'play wagner':", tokens)
	}

	tokens = reSoundBoard.FindStringSubmatch("play wagner 15")
	if tokens[1] != "wagner" || tokens[2] != "15" {
		t.Fatal("couldn't match 'play wagner 15':", tokens)
	}

	tokens = reSoundBoard.FindStringSubmatch("play wagner for 15s")
	if tokens[1] != "wagner" || tokens[2] != "15" {
		t.Fatal("couldn't match 'play wagner for 15s':", tokens)
	}
}

func TestReStop(t *testing.T) {
	if ! reStop.MatchString("stop") {
		t.Fatal("didn't match \"stop\"")
	}

	if ! reStop.MatchString("stahp") {
		t.Fatal("didn't match \"stahp\"")
	}

	if ! reStop.MatchString("stahp") {
		t.Fatal("didn't match \"stahp\"")
	}
}
