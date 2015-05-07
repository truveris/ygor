// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func TestModulePingSend(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	buf := bytes.NewBuffer(nil)
	log.SetOutput(buf)

	m := &PingModule{}
	m.Init(srv)
	m.PrivMsg(srv, &Message{
		ReplyTo: "#test",
		Body:    "whygor: ping",
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 2)
	AssertStringEquals(t, msgs[0].QueueURL, "http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi1")
	AssertStringHasPrefix(t, msgs[0].Body, "ping ")
	AssertStringEquals(t, msgs[1].QueueURL, "http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi2")
	AssertStringHasPrefix(t, msgs[1].Body, "ping ")

	m.MinionMsg(srv, &Message{
		UserID: "UserID-pi1",
		Args:   []string{"1234567890000000000"},
	})

	msgs = srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 0)
	AssertStringContains(t, buf.String(), "got old ping response")

	log.SetOutput(os.Stdout)
}
