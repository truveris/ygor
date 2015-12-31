// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModulePlayUsageOnNoParams(t *testing.T) {
	srv := CreateTestServer()
	client := srv.RegisterClient("dummy", "#test")

	m := &PlayModule{}
	m.Init(srv)
	m.PrivMsg(srv, &InputMessage{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "usage: play url [end]", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}
