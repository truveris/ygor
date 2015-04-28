// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"
)

func TestModuleShutup(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &ShutUpModule{}
	m.Init(srv)
	m.PrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 3)
	AssertStringEquals(t, msgs[0].QueueURL, "http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi1")
	AssertStringEquals(t, msgs[0].Body, "shutup")
	AssertStringEquals(t, msgs[1].QueueURL, "http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi2")
	AssertStringEquals(t, msgs[1].Body, "shutup")
	AssertStringEquals(t, msgs[2].Channel, "#test")
	AssertStringEquals(t, msgs[2].Body, "ok...")
}
