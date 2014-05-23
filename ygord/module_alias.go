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

	"github.com/truveris/ygor"
)

const (
	// That should be plenty for most IRC servers to handle.
	MaxCharsPerPage = 444
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

	if alias == nil {
		Aliases.Add(name, strings.Join(msg.Args[1:], " "))
		outputMsg = "ok (created)"
	} else {
		alias.Value = strings.Join(msg.Args[1:], " ")
		outputMsg = "ok (replaced)"
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
	sort.Strings(aliases)
	first := true
	for _, page := range getPagesOfAliases(aliases) {
		if first {
			IRCPrivMsg(msg.ReplyTo, "known aliases: "+page)
			first = false
		} else {
			IRCPrivMsg(msg.ReplyTo, "... "+page)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (module *AliasModule) GrepCmdFunc(msg *ygor.Message) {
	if len(msg.Args) != 1 && msg.Args[0] != "" {
		IRCPrivMsg(msg.ReplyTo, "usage: grep pattern")
		return
	}

	results := make([]string, 0)
	aliases := Aliases.Names()
	sort.Strings(aliases)
	for _, name := range aliases {
		if strings.Contains(name, msg.Args[0]) {
			results = append(results, name)
		}
	}

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
