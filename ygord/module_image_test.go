// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"
)

func TestModuleImage_Usage(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &ImageModule{}
	m.Init(srv)
	m.PrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "usage: image url")
}

func TestModuleImage_Normal(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &ImageModule{}
	m.Init(srv)
	m.PrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"http://example.com/image.jpg"},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 2)
	AssertStringEquals(t, msgs[0].QueueURL, "http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi1")
	AssertStringEquals(t, msgs[0].Body, "xombrero open https://truveris.github.io/fullscreen-image/?http://example.com/image.jpg")
	AssertStringEquals(t, msgs[1].QueueURL, "http://sqs.us-east-1.amazonaws.com/000000000000/minion-pi2")
	AssertStringEquals(t, msgs[1].Body, "xombrero open https://truveris.github.io/fullscreen-image/?http://example.com/image.jpg")
}
