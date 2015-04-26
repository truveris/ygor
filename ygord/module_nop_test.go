// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"
)

func TestModuleNop(t *testing.T) {
	srv := &Server{}

	msg := &Message{}

	m := &NopModule{}
	m.Init(srv)

	m.PrivMsg(srv, msg)
}
