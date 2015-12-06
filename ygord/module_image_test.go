// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"
)

func TestModuleImage_Usage(t *testing.T) {
	srv := CreateTestServer(t)

	m := &ImageModule{}
	m.Init(srv)
	m.PrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{},
	})

	//	msgs := srv.FlushOutputQueue()
	//	AssertIntEquals(t, len(msgs), 1)
	//	AssertStringEquals(t, msgs[0].Channel, "#test")
	//	AssertStringEquals(t, msgs[0].Body, "usage: image url")
}
