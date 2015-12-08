// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTestServerClientAndAliasModule() (*Server, *Client, *AliasModule) {
	srv := CreateTestServer()
	client := srv.RegisterClient("dummy", "#test")

	module := &AliasModule{}
	module.Init(srv)

	return srv, client, module
}

func TestModuleAliasUsageOnNoParams(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	module.AliasPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "usage: alias name [expr ...]", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleAliasValueNotFound(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	module.AliasPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"key"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "error: unknown alias", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleAliasValueFound(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	srv.Aliases.Add("key", "value", "human", fakeNow)

	module.AliasPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"key"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "key=\"value\" (created by human on 1982-10-20T19:00:00Z)", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleAliasValueFoundWithPercent(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	srv.Aliases.Add("60%", "value", "human", fakeNow)

	module.AliasPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"60%"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "60%=\"value\" (created by human on 1982-10-20T19:00:00Z)", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleAliasValueFoundNested(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	srv.Aliases.Add("key", "value", "human", fakeNow)
	srv.Aliases.Add("value", "null", "robot", fakeNow)

	module.AliasPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"key"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "key=\"value\" (created by human on 1982-10-20T19:00:00Z)", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleAliasChange(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	srv.Aliases.Add("key", "value", "human", fakeNow)

	module.AliasPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"key", "other value"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "ok (replaces \"value\")", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleAliasCreate(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	module.AliasPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"key", "value"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "ok (created)", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleAliasCreateIncremental(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	module.AliasPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"key#", "value1"},
	})
	module.AliasPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"key#", "value2"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 2) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "ok (created as \"key1\")", msgs[0].Body)
		assert.Equal(t, "#test", msgs[1].Channel)
		assert.Equal(t, "ok (created as \"key2\")", msgs[1].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleAliasCreateIncrementalDupeValue(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	module.AliasPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"key#", "value"},
	})
	module.AliasPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"key#", "value"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 2) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "ok (created as \"key1\")", msgs[0].Body)
		assert.Equal(t, "#test", msgs[1].Channel)
		assert.Equal(t, "error: already exists as \"key1\"", msgs[1].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleAliasCreateIncrementalTooManyHashes(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	module.AliasPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"key##", "value"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "error: too many '#'", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleAliases(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	srv.Aliases.Add("foo", "foo-value", "human", fakeNow)
	srv.Aliases.Add("bar", "bar-value", "human", fakeNow)
	srv.Aliases.Add("zzz", "zzz-value", "human", fakeNow)

	module.AliasesPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "known aliases: bar, foo, zzz", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleAliasesPages(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	for i := 0; i < 35; i++ {
		key := fmt.Sprintf("foobarfoobar%d", i)
		srv.Aliases.Add(key, "foo-value", "human", fakeNow)
	}

	module.AliasesPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 2) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "known aliases: foobarfoobar0, foobarfoobar1, foobarfoobar10, foobarfoobar11, foobarfoobar12, foobarfoobar13, foobarfoobar14, foobarfoobar15, foobarfoobar16, foobarfoobar17, foobarfoobar18, foobarfoobar19, foobarfoobar2, foobarfoobar20, foobarfoobar21, foobarfoobar22, foobarfoobar23, foobarfoobar24, foobarfoobar25, foobarfoobar26, foobarfoobar27, foobarfoobar28, foobarfoobar29, foobarfoobar3, foobarfoobar30, foobarfoobar31, foobarfoobar32, foobarfoobar33", msgs[0].Body)
		assert.Equal(t, "#test", msgs[1].Channel)
		assert.Equal(t, "... foobarfoobar34, foobarfoobar4, foobarfoobar5, foobarfoobar6, foobarfoobar7, foobarfoobar8, foobarfoobar9", msgs[1].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleAliasesTooMany(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("foobarfoobar%d", i)
		srv.Aliases.Add(key, "foo-value", "human", fakeNow)
	}

	module.AliasesPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "error: too many results, use grep", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleUnaliasUsage(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	srv.Aliases.Add("foo", "foo-value", "human", fakeNow)

	module.UnAliasPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "usage: unalias name", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleUnaliasExisting(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	srv.Aliases.Add("foo", "foo-value", "human", fakeNow)

	module.UnAliasPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"foo"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "ok (deleted)", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())

	assert.Nil(t, srv.Aliases.Get("foo"))
}

func TestModuleUnaliasNonExisting(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	srv.Aliases.Add("foo", "foo-value", "human", fakeNow)

	module.UnAliasPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"bar"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "error: unknown alias", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleGrepUsage(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	module.GrepPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "usage: grep pattern", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleGrep(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	srv.Aliases.Add("foo", "foo-value", "human", fakeNow)
	srv.Aliases.Add("baz", "baz-value", "human", fakeNow)
	srv.Aliases.Add("zzz", "zzz-value", "human", fakeNow)

	module.GrepPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"z"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "baz, zzz", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleGrepNoResult(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	srv.Aliases.Add("foo", "foo-value", "human", fakeNow)
	srv.Aliases.Add("bar", "bar-value", "human", fakeNow)

	module.GrepPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"z"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "error: no matches found", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleGrepTooManyResults(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("foo%d", i)
		srv.Aliases.Add(key, "foo-value", "human", fakeNow)
	}

	module.GrepPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"foo"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "error: too many matches, refine your search", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleRandom(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	srv.Aliases.Add("foo", "play foo.mp3", "human", fakeNow)
	srv.Aliases.Add("bar", "play bar.mp3", "human", fakeNow)
	srv.Aliases.Add("zzz", "play zzz.mp3", "human", fakeNow)

	module.RandomPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"o"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 2) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "chooses foo", msgs[0].Body)
		assert.Equal(t, "#test", msgs[1].Channel)
		assert.Equal(t, "command not found: play", msgs[1].Body)
	}

	assert.Empty(t, client.FlushQueue())
}

func TestModuleRandomNotFound(t *testing.T) {
	srv, client, module := createTestServerClientAndAliasModule()

	srv.Aliases.Add("foo", "play foo.mp3", "human", fakeNow)
	srv.Aliases.Add("bar", "play bar.mp3", "human", fakeNow)

	module.RandomPrivMsg(srv, &IRCInputMessage{
		ReplyTo: "#test",
		Args:    []string{"w"},
	})

	msgs := srv.FlushIRCOutputQueue()
	if assert.Len(t, msgs, 1) {
		assert.Equal(t, "#test", msgs[0].Channel)
		assert.Equal(t, "error: no matches found", msgs[0].Body)
	}

	assert.Empty(t, client.FlushQueue())
}
