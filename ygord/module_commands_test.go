// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModuleCommands(t *testing.T) {
	srv := CreateTestServer()
	client := srv.GetClientFromID(srv.RegisterClient("dummy", "#test"))

	// Register a module so it shows up in the listing.
	(&NopModule{}).Init(srv)

	m := &CommandsModule{}
	m.Init(srv)
	m.PrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "commands, nop", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}
