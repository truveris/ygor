// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModuleNop(t *testing.T) {
	srv := CreateTestServer()
	client := srv.RegisterClient("dummy", "#test")

	m := &NopModule{}
	m.Init(srv)
	m.PrivMsg(srv, &InputMessage{ReplyTo: "#test"})

	assert.Empty(t, srv.FlushIRCOutputQueue())
	assert.Empty(t, client.FlushQueue())
}
