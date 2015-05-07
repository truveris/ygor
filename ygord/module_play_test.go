// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"
)

func TestModulePlayUsageOnNoParams(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &PlayModule{}
	m.Init(srv)
	m.PrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "usage: play url [duration]")
}

func TestModulePlayStuff(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &PlayModule{}
	m.Init(srv)
	m.PrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"stuff.ogg"},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 2)
	AssertStringEquals(t, msgs[0].QueueURL, "http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi1")
	AssertStringEquals(t, msgs[0].Body, "play stuff.ogg")
	AssertStringEquals(t, msgs[1].QueueURL, "http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi2")
	AssertStringEquals(t, msgs[1].Body, "play stuff.ogg")
}

func TestModulePlayStuffWithDuration(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &PlayModule{}
	m.Init(srv)
	m.PrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"stuff.ogg", "5s"},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 2)
	AssertStringEquals(t, msgs[0].QueueURL, "http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi1")
	AssertStringEquals(t, msgs[0].Body, "play stuff.ogg 5s")
	AssertStringEquals(t, msgs[1].QueueURL, "http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi2")
	AssertStringEquals(t, msgs[1].Body, "play stuff.ogg 5s")
}
