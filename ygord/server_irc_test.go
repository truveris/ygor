// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thoj/go-ircevent"
)

func TestServerIRCAddressedWithoutSpace(t *testing.T) {
	srv := CreateTestServer()

	msgs := srv.NewMessagesFromIRCEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "whygore:minions"},
	})

	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].ReplyTo)
		assert.Equal(t, "minions", msgs[0].Body)
	}
}

func TestServerIRCAddressedWithoutColon(t *testing.T) {
	srv := CreateTestServer()

	msgs := srv.NewMessagesFromIRCEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "whygore minions"},
	})

	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].ReplyTo)
		assert.Equal(t, "minions", msgs[0].Body)
	}
}

func TestServerIRCAddressedSpacesEverywhere(t *testing.T) {
	srv := CreateTestServer()

	msgs := srv.NewMessagesFromIRCEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "  whygore   minions  "},
	})

	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].ReplyTo)
		assert.Equal(t, "minions", msgs[0].Body)
	}
}

func TestServerIRCResolveAlias(t *testing.T) {
	srv := CreateTestServer()
	srv.Aliases.Add("coffee", "play freshpots.mp3", "human", fakeNow)

	msgs := srv.NewMessagesFromIRCEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "whygore: coffee"},
	})

	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].ReplyTo)
		assert.Equal(t, "play freshpots.mp3", msgs[0].Body)
	}
}

func TestServerIRCSeparateStatements(t *testing.T) {
	srv := CreateTestServer()

	msgs := srv.NewMessagesFromIRCEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "whygore: play freshpots.mp3; image freshpots.gif"},
	})

	if assert.Len(t, msgs, 2) {
		assert.Equal(t, "#test", msgs[0].ReplyTo)
		assert.Equal(t, "play freshpots.mp3", msgs[0].Body)
		assert.Equal(t, "#test", msgs[1].ReplyTo)
		assert.Equal(t, "image freshpots.gif", msgs[1].Body)
	}
}

func TestServerIRCSeparateStatementsInAlias(t *testing.T) {
	srv := CreateTestServer()
	srv.Aliases.Add("coffee", "play freshpots.mp3; image freshpots.gif", "human", fakeNow)

	msgs := srv.NewMessagesFromIRCEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "whygore: coffee"},
	})

	if assert.Len(t, msgs, 2) {
		assert.Equal(t, "#test", msgs[0].ReplyTo)
		assert.Equal(t, "play freshpots.mp3", msgs[0].Body)
		assert.Equal(t, "#test", msgs[1].ReplyTo)
		assert.Equal(t, "image freshpots.gif", msgs[1].Body)
	}
}

func TestServerIRCSeparateStatementsInNestedAlias(t *testing.T) {
	srv := CreateTestServer()
	srv.Aliases.Add("coffee.mp3", "play freshpots.mp3; nop", "human", fakeNow)
	srv.Aliases.Add("coffee.gif", "image freshpots.gif; nop", "human", fakeNow)
	srv.Aliases.Add("coffee", "coffee.mp3; coffee.gif", "human", fakeNow)

	msgs := srv.NewMessagesFromIRCEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "whygore: coffee"},
	})

	if assert.Len(t, msgs, 4) {
		assert.Equal(t, "#test", msgs[0].ReplyTo)
		assert.Equal(t, "play freshpots.mp3", msgs[0].Body)
		assert.Equal(t, "#test", msgs[1].ReplyTo)
		assert.Equal(t, "nop", msgs[1].Body)
		assert.Equal(t, "#test", msgs[2].ReplyTo)
		assert.Equal(t, "image freshpots.gif", msgs[2].Body)
		assert.Equal(t, "#test", msgs[3].ReplyTo)
		assert.Equal(t, "nop", msgs[3].Body)
	}
}

func TestServerIRCMaxRecursionAlias(t *testing.T) {
	srv := CreateTestServer()
	srv.Aliases.Add("coffee", "bean", "human", fakeNow)
	srv.Aliases.Add("bean", "coffee", "human", fakeNow)

	msgs := srv.NewMessagesFromIRCEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "whygore: coffee"},
	})

	assert.Empty(t, msgs)

	omsgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, omsgs, 1) {
		assert.Equal(t, "#test", omsgs[0].Channel)
		assert.Equal(t, "lexer/expand error: max recursion reached", omsgs[0].Body)
	}
}

func TestServerIRCMaxRecursionRandom(t *testing.T) {
	srv := CreateTestServer()
	srv.Aliases.Add("coffee", "bean", "human", fakeNow)
	srv.Aliases.Add("bean", "random coffee", "human", fakeNow)

	msgs := srv.NewMessagesFromIRCEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "whygore: coffee"},
	})

	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].ReplyTo)
		assert.Equal(t, "random coffee", msgs[0].Body)
		assert.Equal(t, 1, msgs[0].Recursion)
	}

}
