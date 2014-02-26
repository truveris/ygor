// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"
)

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
