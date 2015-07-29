// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"testing"
	"time"
)

var (
	now = time.Date(1982, 10, 20, 16, 0, 0, 0, time.UTC)
)

func TestModuleAliasUsageOnNoParams(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &AliasModule{}
	m.Init(srv)
	m.AliasPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "usage: alias name [expr ...]")
}

func TestModuleAliasValueNotFound(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &AliasModule{}
	m.Init(srv)

	m.AliasPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"key"},
	})
	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "error: unknown alias")
}

func TestModuleAliasValueFound(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("key", "value", "human", now)

	m := &AliasModule{}
	m.Init(srv)
	m.AliasPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"key"},
	})
	msgs := srv.FlushOutputQueue()

	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "key=\"value\" (created by human on 1982-10-20T16:00:00Z)")
}

func TestModuleAliasValueFoundWithPercent(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("60%", "value", "human", now)

	m := &AliasModule{}
	m.Init(srv)
	m.AliasPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"60%"},
	})
	msgs := srv.FlushOutputQueue()

	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "60%=\"value\" (created by human on 1982-10-20T16:00:00Z)")
}

func TestModuleAliasValueFoundNested(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("key", "value", "human", now)
	srv.Aliases.Add("value", "null", "robot", now)

	m := &AliasModule{}
	m.Init(srv)
	m.AliasPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"key"},
	})
	msgs := srv.FlushOutputQueue()

	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "key=\"value\" (created by human on 1982-10-20T16:00:00Z)")
}

func TestModuleAliasChange(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("key", "value", "human", now)

	m := &AliasModule{}
	m.Init(srv)
	m.AliasPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"key", "other value"},
	})
	msgs := srv.FlushOutputQueue()

	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "ok (replaces \"value\")")
}

func TestModuleAliasCreate(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &AliasModule{}
	m.Init(srv)
	m.AliasPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"key", "value"},
	})
	msgs := srv.FlushOutputQueue()

	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "ok (created)")
}

func TestModuleAliasCreateIncremental(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &AliasModule{}
	m.Init(srv)
	m.AliasPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"key#", "value1"},
	})
	m.AliasPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"key#", "value2"},
	})
	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 2)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "ok (created as \"key1\")")
	AssertStringEquals(t, msgs[1].Channel, "#test")
	AssertStringEquals(t, msgs[1].Body, "ok (created as \"key2\")")
}

func TestModuleAliasCreateIncrementalDupeValue(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &AliasModule{}
	m.Init(srv)
	m.AliasPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"key#", "value"},
	})
	m.AliasPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"key#", "value"},
	})
	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 2)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "ok (created as \"key1\")")
	AssertStringEquals(t, msgs[1].Channel, "#test")
	AssertStringEquals(t, msgs[1].Body, "error: already exists as \"key1\"")
}

func TestModuleAliasCreateIncrementalTooManyHashes(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &AliasModule{}
	m.Init(srv)
	m.AliasPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"key##", "value"},
	})
	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "error: too many '#'")
}

func TestModuleAliases(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("foo", "foo-value", "human", now)
	srv.Aliases.Add("bar", "bar-value", "human", now)
	srv.Aliases.Add("zzz", "zzz-value", "human", now)

	m := &AliasModule{}
	m.Init(srv)
	m.AliasesPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "known aliases: bar, foo, zzz")
}

func TestModuleAliasesPages(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)
	for i := 0; i < 35; i++ {
		key := fmt.Sprintf("foobarfoobar%d", i)
		srv.Aliases.Add(key, "foo-value", "human", now)
	}

	m := &AliasModule{}
	m.Init(srv)
	m.AliasesPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 2)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "known aliases: foobarfoobar0, foobarfoobar1, foobarfoobar10, foobarfoobar11, foobarfoobar12, foobarfoobar13, foobarfoobar14, foobarfoobar15, foobarfoobar16, foobarfoobar17, foobarfoobar18, foobarfoobar19, foobarfoobar2, foobarfoobar20, foobarfoobar21, foobarfoobar22, foobarfoobar23, foobarfoobar24, foobarfoobar25, foobarfoobar26, foobarfoobar27, foobarfoobar28, foobarfoobar29, foobarfoobar3, foobarfoobar30, foobarfoobar31, foobarfoobar32, foobarfoobar33")
	AssertStringEquals(t, msgs[1].Channel, "#test")
	AssertStringEquals(t, msgs[1].Body, "... foobarfoobar34, foobarfoobar4, foobarfoobar5, foobarfoobar6, foobarfoobar7, foobarfoobar8, foobarfoobar9")
}

func TestModuleAliasesTooMany(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("foobarfoobar%d", i)
		srv.Aliases.Add(key, "foo-value", "human", now)
	}

	m := &AliasModule{}
	m.Init(srv)
	m.AliasesPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "error: too many results, use grep")
}

func TestModuleUnaliasUsage(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("foo", "foo-value", "human", now)

	m := &AliasModule{}
	m.Init(srv)
	m.UnAliasPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "usage: unalias name")
}

func TestModuleUnaliasExisting(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("foo", "foo-value", "human", now)

	m := &AliasModule{}
	m.Init(srv)
	m.UnAliasPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"foo"},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "ok (deleted)")

	a := srv.Aliases.Get("foo")
	if a != nil {
		t.Error("failed to delete alias")
	}
}

func TestModuleUnaliasNonExisting(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("foo", "foo-value", "human", now)

	m := &AliasModule{}
	m.Init(srv)
	m.UnAliasPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"bar"},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "error: unknown alias")
}

func TestModuleGrepUsage(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)

	m := &AliasModule{}
	m.Init(srv)
	m.GrepPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "usage: grep pattern")
}

func TestModuleGrep(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("foo", "foo-value", "human", now)
	srv.Aliases.Add("baz", "baz-value", "human", now)
	srv.Aliases.Add("zzz", "zzz-value", "human", now)

	m := &AliasModule{}
	m.Init(srv)
	m.GrepPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"z"},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "baz, zzz")
}

func TestModuleGrepNoResult(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("foo", "foo-value", "human", now)
	srv.Aliases.Add("bar", "bar-value", "human", now)

	m := &AliasModule{}
	m.Init(srv)
	m.GrepPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"z"},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "error: no matches found")
}

func TestModuleGrepTooManyResults(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("foo%d", i)
		srv.Aliases.Add(key, "foo-value", "human", now)
	}

	m := &AliasModule{}
	m.Init(srv)
	m.GrepPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"foo"},
	})

	msgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(msgs), 1)
	AssertStringEquals(t, msgs[0].Channel, "#test")
	AssertStringEquals(t, msgs[0].Body, "error: too many matches, refine your search")
}

func TestModuleRandom(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("foo", "play foo.mp3", "human", now)
	srv.Aliases.Add("bar", "play bar.mp3", "human", now)
	srv.Aliases.Add("zzz", "play zzz.mp3", "human", now)

	m := &AliasModule{}
	m.Init(srv)
	m.RandomPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"o"},
	})

	omsgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(omsgs), 2)
	AssertStringEquals(t, omsgs[0].Channel, "#test")
	AssertStringEquals(t, omsgs[0].Body, "chooses foo")

	// play is not found because unregistered.
	AssertStringEquals(t, omsgs[1].Channel, "#test")
	AssertStringEquals(t, omsgs[1].Body, "command not found: play")
}

func TestModuleRandomNotFound(t *testing.T) {
	srv := CreateTestServerWithTwoMinions(t)
	srv.Aliases.Add("foo", "play foo.mp3", "human", now)
	srv.Aliases.Add("bar", "play bar.mp3", "human", now)

	m := &AliasModule{}
	m.Init(srv)
	m.RandomPrivMsg(srv, &Message{
		ReplyTo: "#test",
		Args:    []string{"w"},
	})

	omsgs := srv.FlushOutputQueue()
	AssertIntEquals(t, len(omsgs), 1)
	AssertStringEquals(t, omsgs[0].Channel, "#test")
	AssertStringEquals(t, omsgs[0].Body, "error: no matches found")
}
