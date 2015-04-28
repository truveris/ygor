// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"
)

func TestModuleNop(t *testing.T) {
	srv := CreateServer(&Config{
		AliasFilePath:   "/dev/null",
		MinionsFilePath: "/dev/null",
	})

	m := &NopModule{}
	m.Init(srv)
	m.PrivMsg(srv, &Message{})

	msgs := srv.FlushOutputQueue()
	if len(msgs) != 0 {
		t.Error("Outgoing message queue should be empty: ", len(msgs))
	}
}
