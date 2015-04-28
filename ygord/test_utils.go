// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"testing"
)

func AssertStringEquals(t *testing.T, a, b string) {
	if a != b {
		t.Error(fmt.Sprintf("Strings to do not match (a=%s, b=%s)", a, b))
	}
}

func AssertIntEquals(t *testing.T, a, b int) {
	if a != b {
		t.Error(fmt.Sprintf("Integers to do not match (a=%d, b=%d)", a, b))
	}
}

func RegisterTestMinion(t *testing.T, srv *Server, name string) {
	// Register a minion
	rm := &MinionsModule{}
	rm.MinionMsg(srv, &Message{
		UserID: "UserID-" + name,
		Args:   []string{name, "http://sqs.us-east-1.amazonaws.com/000000000000/minion-" + name},
	})
	msgs := srv.FlushOutputQueue()
	if len(msgs) != 1 {
		t.Error("Outgoing message queue should have one message, not ", len(msgs))
		return
	}
	AssertStringEquals(t, msgs[0].QueueURL, "http://sqs.us-east-1.amazonaws.com/000000000000/minion-"+name)
	AssertStringEquals(t, msgs[0].Body, "register success")
}

func CreateTestServerWithTwoMinions(t *testing.T) *Server {
	srv := CreateServer(&Config{
		IRCNickname:     "whygore",
		AliasFilePath:   ":memory:",
		MinionsFilePath: ":memory:",
		Channels: map[string]ChannelCfg{
			"#test": ChannelCfg{
				Minions: []string{"pi1", "pi2"},
			},
		},
	})

	RegisterTestMinion(t, srv, "pi1")
	RegisterTestMinion(t, srv, "pi2")

	return srv
}
