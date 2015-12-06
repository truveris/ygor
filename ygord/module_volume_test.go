// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModuleVolume_UsageNoParams(t *testing.T) {
	srv := CreateTestServer()
	client := srv.GetClientFromID(srv.RegisterClient("dummy", "#test"))

	m := &VolumeModule{}
	m.Init(srv)
	m.PrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "usage: volume percent", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleVolume_BadValues(t *testing.T) {
	srv := CreateTestServer()
	client := srv.GetClientFromID(srv.RegisterClient("dummy", "#test"))

	m := &VolumeModule{}
	m.Init(srv)

	m.PrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"-10%"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "error: bad input, must be absolute rounded percent value (e.g. 42%)", msgs[0].Body)
	}
	assert.Empty(t, client.FlushQueue())

	m.PrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"wat"},
	})

	msgs = srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "error: bad input, must be absolute rounded percent value (e.g. 42%)", msgs[0].Body)
	}
	assert.Empty(t, client.FlushQueue())
}
