// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"github.com/thoj/go-ircevent"
	"testing"
	"time"
)

func TestServerIRCAddressedWithoutSpace(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	msgs := srv.NewMessagesFromEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "whygore:minions"},
	})

	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Body, "minions")
}

func TestServerIRCAddressedWithoutColon(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	msgs := srv.NewMessagesFromEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "whygore minions"},
	})

	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Body, "minions")
}

func TestServerIRCAddressedSpacesEverywhere(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	msgs := srv.NewMessagesFromEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "  whygore   minions  "},
	})

	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Body, "minions")
}

func TestServerIRCResolveAlias(t *testing.T) {
	now := time.Date(1982, 10, 20, 16, 0, 0, 0, time.UTC)
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("coffee", "play freshpots.mp3", "human", now)

	msgs := srv.NewMessagesFromEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "whygore: coffee"},
	})

	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Body, "play freshpots.mp3")
}

func TestServerIRCSeparateStatements(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	msgs := srv.NewMessagesFromEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "whygore: play freshpots.mp3; image freshpots.gif"},
	})

	AssertIntEquals(t, len(msgs), 2)
	AssertStringEquals(t, msgs[0].Body, "play freshpots.mp3")
	AssertStringEquals(t, msgs[1].Body, "image freshpots.gif")
}

func TestServerIRCSeparateStatementsInAlias(t *testing.T) {
	now := time.Date(1982, 10, 20, 16, 0, 0, 0, time.UTC)
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("coffee", "play freshpots.mp3; image freshpots.gif", "human", now)

	msgs := srv.NewMessagesFromEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "whygore: coffee"},
	})

	AssertIntEquals(t, len(msgs), 2)
	AssertStringEquals(t, msgs[0].Body, "play freshpots.mp3")
	AssertStringEquals(t, msgs[1].Body, "image freshpots.gif")
}

func TestServerIRCSeparateStatementsInNestedAlias(t *testing.T) {
	now := time.Date(1982, 10, 20, 16, 0, 0, 0, time.UTC)
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("coffee.mp3", "play freshpots.mp3; nop", "human", now)
	srv.Aliases.Add("coffee.gif", "image freshpots.gif; nop", "human", now)
	srv.Aliases.Add("coffee", "coffee.mp3; coffee.gif", "human", now)

	msgs := srv.NewMessagesFromEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "whygore: coffee"},
	})

	AssertIntEquals(t, len(msgs), 4)
	AssertStringEquals(t, msgs[0].Body, "play freshpots.mp3")
	AssertStringEquals(t, msgs[1].Body, "nop")
	AssertStringEquals(t, msgs[2].Body, "image freshpots.gif")
	AssertStringEquals(t, msgs[3].Body, "nop")
}

func TestServerIRCMaxRecursionAlias(t *testing.T) {
	now := time.Date(1982, 10, 20, 16, 0, 0, 0, time.UTC)
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("coffee", "bean", "human", now)
	srv.Aliases.Add("bean", "coffee", "human", now)

	msgs := srv.NewMessagesFromEvent(&irc.Event{
		Code:      "PRIVMSG",
		Nick:      "foobar",
		Arguments: []string{"#test", "whygore: coffee"},
	})

	AssertIntEquals(t, len(msgs), 0)

	omsgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(omsgs), 1)
	AssertStringEquals(t, omsgs[0].Channel, "#test")
	AssertStringEquals(t, omsgs[0].Body, "lexer/expand error: max recursion reached")
}
