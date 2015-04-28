// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"
)

func TestModuleXombrero_Usage(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &XombreroModule{}
	m.Init(srv)
	m.PrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "usage: xombrero [command [param ...]]")
}

func TestModuleXombrero_Open(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &XombreroModule{}
	m.Init(srv)
	m.PrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"open", "http://example.com/"},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 2)
	AssertStringEquals(t, msgs[0].QueueURL, "http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi1")
	AssertStringEquals(t, msgs[0].Body, "xombrero open http://example.com/")
	AssertStringEquals(t, msgs[1].QueueURL, "http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi2")
	AssertStringEquals(t, msgs[1].Body, "xombrero open http://example.com/")
}

func TestModuleXombrero_MinionError(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &XombreroModule{}
	m.Init(srv)
	m.MinionMsg(srv, &Message{
		UserID: "UserID-pi1",
		Args:   []string{"error", "foo"},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "xombrero@pi1: error foo")
}

func TestModuleXombrero_MinionAck(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &XombreroModule{}
	m.Init(srv)
	m.MinionMsg(srv, &Message{
		UserID: "UserID-pi1",
		Args:   []string{"ok"},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 0)
}
