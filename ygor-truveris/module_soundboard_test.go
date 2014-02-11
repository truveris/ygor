// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"
)

func TestReSoundBoard(t *testing.T) {
	tokens := reSoundBoard.FindStringSubmatch("anything")
	if tokens == nil {
		t.Fatal("couldn't match 'anything'")
	}

	tokens = reSoundBoard.FindStringSubmatch("wagner")
	if tokens[1] != "wagner"{
		t.Fatal("couldn't match 'wagner' by itself:", tokens)
	}

	tokens = reSoundBoard.FindStringSubmatch("blast wagner")
	if tokens[1] != "wagner"{
		t.Fatal("couldn't match 'blast wagner':", tokens)
	}

	tokens = reSoundBoard.FindStringSubmatch("blast wagner 15")
	if tokens[1] != "wagner" || tokens[2] != "15" {
		t.Fatal("couldn't match 'blast wagner 15':", tokens)
	}

	tokens = reSoundBoard.FindStringSubmatch("blast wagner for 15s")
	if tokens[1] != "wagner" || tokens[2] != "15" {
		t.Fatal("couldn't match 'blast wagner for 15s':", tokens)
	}
}

func TestReAddressed(t *testing.T) {
	tokens := reAddressed.FindStringSubmatch("ygor, you're a moron")
	if tokens[1] != "ygor" || tokens[2] != "you're a moron" {
		t.Fatal("couldn't match 'ygor, you're a moron'")
	}

	tokens = reAddressed.FindStringSubmatch("ygor... get a life")
	if tokens[1] != "ygor" || tokens[2] != "get a life" {
		t.Fatal("couldn't match 'ygor... get a life'")
	}

	tokens = reAddressed.FindStringSubmatch("ygorz!")
	if tokens != nil {
		t.Fatal("matched 'ygorz!' by mistake'")
	}
}
