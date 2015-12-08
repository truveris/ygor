// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModuleSayUsageOnNoParams(t *testing.T) {
	srv := CreateTestServer()
	client := srv.RegisterClient("dummy", "#test")

	m := &SayModule{}
	m.Init(srv)
	m.PrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "usage: say [-v voice] sentence", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleSayNoConfig(t *testing.T) {
	srv := CreateTestServer()
	client := srv.RegisterClient("dummy", "#test")

	m := &SayModule{}
	m.Init(srv)
	m.PrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"hello"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "error: SaydURL is not configured", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}
