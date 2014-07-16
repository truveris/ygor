// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This module allows channel users to configure aliases themselves.

package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"math/rand"

	"github.com/truveris/ygor"
)

const (
	// That should be plenty for most IRC servers to handle.
	MaxCharsPerPage = 444

	// You can't list that many aliases without trouble...
	MaxAliasesForFullList = 40
)

type AliasModule struct{}

func (module AliasModule) PrivMsg(msg *ygor.PrivMsg) {}

// Command used to set a new alias.
func (module *AliasModule) AliasCmdFunc(msg *ygor.Message) {
	var outputMsg string

	if len(msg.Args) == 0 {
		IRCPrivMsg(msg.ReplyTo, "usage: alias name [command [params ...]]")
		return
	}

	name := msg.Args[0]
	alias := Aliases.Get(name)

	// Request the value of an alias.
	if len(msg.Args) == 1 {
		if alias == nil {
			IRCPrivMsg(msg.ReplyTo, "error: unknown alias")
			return
		}
		IRCPrivMsg(msg.ReplyTo, fmt.Sprintf("'%s' is an alias for '%s'",
			alias.Name, alias.Value))
		return
	}

	// Set a new alias.
	cmd := ygor.GetCommand(name)
	if cmd != nil {
		IRCPrivMsg(msg.ReplyTo, fmt.Sprintf("error: '%s' is already a"+
			" command", name))
		return
	}

	cmd = ygor.GetCommand(msg.Args[1])
	if cmd == nil {
		IRCPrivMsg(msg.ReplyTo, fmt.Sprintf("error: '%s' is not a valid "+
			"command", msg.Args[1]))
		return
	}

	newValue := strings.Join(msg.Args[1:], " ")

	if alias == nil {
		outputMsg = "ok (created)"
		Aliases.Add(name, newValue)
	} else if alias.Value == newValue {
		outputMsg = "no changes"
	} else {
		outputMsg = "ok (replaces \"" + alias.Value + "\")"
		alias.Value = newValue
	}

	err := Aliases.Save()
	if err != nil {
		outputMsg = "error: " + err.Error()
	}

	IRCPrivMsg(msg.ReplyTo, outputMsg)
}

// Take a list of aliases, return joined pages.
func getPagesOfAliases(aliases []string) []string {
	length := 0
	pages := make([]string, 0)

	for i := 0; i < len(aliases); {
		var page []string

		if length > 0 {
			length += len(", ")
		}

		length += len(aliases[i])

		if length > MaxCharsPerPage {
			page, aliases = aliases[:i], aliases[i:]
			pages = append(pages, strings.Join(page, ", "))
			length = 0
			i = 0
			continue
		}

		i++
	}

	if length > 0 {
		pages = append(pages, strings.Join(aliases, ", "))
	}

	return pages
}

func (module *AliasModule) UnAliasCmdFunc(msg *ygor.Message) {
	if len(msg.Args) != 1 {
		IRCPrivMsg(msg.ReplyTo, "usage: unalias name")
		return
	}

	name := msg.Args[0]
	alias := Aliases.Get(name)

	if alias == nil {
		IRCPrivMsg(msg.ReplyTo, "error: unknown alias")
		return
	} else {
		Aliases.Delete(name)
		IRCPrivMsg(msg.ReplyTo, "ok (deleted)")
	}
	Aliases.Save()
}

func (module *AliasModule) AliasesCmdFunc(msg *ygor.Message) {
	if len(msg.Args) != 0 {
		IRCPrivMsg(msg.ReplyTo, "usage: aliases")
		return
	}

	aliases := Aliases.Names()

	if len(aliases) > MaxAliasesForFullList {
		IRCPrivMsg(msg.ReplyTo, "error: too many results, use grep")
		return
	}

	sort.Strings(aliases)
	first := true
	for _, page := range getPagesOfAliases(aliases) {
		if first {
			IRCPrivMsg(msg.ReplyTo, "known aliases: "+page)
			first = false
		} else {
			IRCPrivMsg(msg.ReplyTo, "... "+page)
		}
		if !cfg.TestMode {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (module *AliasModule) GrepCmdFunc(msg *ygor.Message) {
	if len(msg.Args) != 1 && msg.Args[0] != "" {
		IRCPrivMsg(msg.ReplyTo, "usage: grep pattern")
		return
	}

	results := Aliases.Find(msg.Args[0])
	sort.Strings(results)

	if len(results) == 0 {
		IRCPrivMsg(msg.ReplyTo, "error: no results")
		return
	}

	found := strings.Join(results, ", ")
	if len(found) > MaxCharsPerPage {
		IRCPrivMsg(msg.ReplyTo, "error: too many results, refine your search")
		return
	}

	IRCPrivMsg(msg.ReplyTo, found)
}

func (module *AliasModule) RandomCmdFunc(msg *ygor.Message) {
	var aliases []string

	switch len(msg.Args) {
	case 0:
		aliases = Aliases.Names()
	case 1:
		aliases = Aliases.Find(msg.Args[0])
	default:
		IRCPrivMsg(msg.ReplyTo, "usage: random [pattern]")
		return
	}

	if len(aliases) <= 0 {
		IRCPrivMsg(msg.ReplyTo, "no matches found")
		return
	}

	idx := rand.Intn(len(aliases))

	body, err := Aliases.Resolve(aliases[idx])
	if err != nil {
		Debug("failed to resolve aliases: " + err.Error())
		return
	}

	IRCPrivMsg(msg.ReplyTo, "the codes have chosen "+aliases[idx])

	privmsg := &ygor.PrivMsg{}
	privmsg.Nick = msg.UserID
	privmsg.Body = body
	privmsg.ReplyTo = msg.ReplyTo
	privmsg.Addressed = true
	newmsg := NewMessageFromPrivMsg(privmsg)
	if newmsg == nil {
		Debug("failed to convert PRIVMSG")
		return
	}
	InputQueue <- newmsg
}

func (module *AliasModule) Init() {
	ygor.RegisterCommand(ygor.Command{
		Name:            "alias",
		PrivMsgFunction: module.AliasCmdFunc,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	ygor.RegisterCommand(ygor.Command{
		Name:            "grep",
		PrivMsgFunction: module.GrepCmdFunc,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	ygor.RegisterCommand(ygor.Command{
		Name:            "random",
		PrivMsgFunction: module.RandomCmdFunc,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	ygor.RegisterCommand(ygor.Command{
		Name:            "unalias",
		PrivMsgFunction: module.UnAliasCmdFunc,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	ygor.RegisterCommand(ygor.Command{
		Name:            "aliases",
		PrivMsgFunction: module.AliasesCmdFunc,
		Addressed:       true,
		AllowPrivate:    true,
		AllowChannel:    true,
	})
}
