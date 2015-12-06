// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestModuleNop(t *testing.T) {
	srv := CreateTestServer(t)
	client := srv.GetClientFromID(srv.RegisterClient("dummy", "#test"))

	m := &NopModule{}
	m.Init(srv)
	m.PrivMsg(srv, &IRCInputMessage{ReplyTo: "#test"})

	assert.Empty(t, srv.FlushIRCOutputQueue())
	assert.Empty(t, client.FlushQueue())
}
