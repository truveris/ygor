// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"
)

func TestModuleVolume_UsageNoParams(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &VolumeModule{}
	m.Init(srv)
	m.PrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "usage: volume percent")
}

func TestModuleVolume_BadValues(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &VolumeModule{}
	m.Init(srv)

	m.PrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"-10%"},
	})
	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "error: bad input, must be absolute rounded percent value (e.g. 42%)")

	m.PrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"wat"},
	})
	msgs = srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "error: bad input, must be absolute rounded percent value (e.g. 42%)")
}

func TestModuleVolume_MinionError(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &VolumeModule{}
	m.Init(srv)

	m.MinionMsg(srv, &Message{
		UserID: "UserID-pi1",
		Args:   []string{"error", "things"},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "volume@pi1: error things")
}
