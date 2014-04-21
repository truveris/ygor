// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This module allows channel users to configure aliases themselves.

package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/truveris/ygor"
)

const (
	// That should be plenty for most IRC servers to handle.
	MaxCharsPerPage = 444
)

type AliasModule struct{}

func (module AliasModule) PrivMsg(msg *ygor.PrivMsg) {}

// Command used to set a new alias.
func AliasCmdFunc(msg *ygor.Message) {
	var outputMsg string

	if len(msg.Args) == 0 {
		IRCPrivMsg(msg.ReplyTo, "usage: alias name [command [params ...]]")
		return
	}

	name := msg.Args[0]
	alias := ygor.GetAlias(name)

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
		ygor.AddAlias(name, strings.Join(msg.Args[1:], " "))
		outputMsg = "ok (created)"
	} else {
		alias.Value = strings.Join(msg.Args[1:], " ")
		outputMsg = "ok (replaced)"
	}

	err := ygor.SaveAliases()
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

func UnAliasCmdFunc(msg *ygor.Message) {
	if len(msg.Args) != 1 {
		IRCPrivMsg(msg.ReplyTo, "usage: unalias name")
		return
	}

	name := msg.Args[0]
	alias := ygor.GetAlias(name)

	if alias == nil {
		IRCPrivMsg(msg.ReplyTo, "error: unknown alias")
		return
	} else {
		ygor.DeleteAlias(name)
		IRCPrivMsg(msg.ReplyTo, "ok (deleted)")
	}
	ygor.SaveAliases()
}

func AliasesCmdFunc(msg *ygor.Message) {
	var aliases []string

	if len(msg.Args) != 0 {
		IRCPrivMsg(msg.ReplyTo, "usage: aliases")
		return
	}

	for _, alias := range ygor.Aliases {
		aliases = append(aliases, alias.Name)
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
	}
}

func (module AliasModule) Init() {
	if cfg.AliasFilePath != "" {
		ygor.SetAliasFilePath(cfg.AliasFilePath)
	}

	ygor.RegisterCommand(ygor.Command{
		Name:            "alias",
		PrivMsgFunction: AliasCmdFunc,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	ygor.RegisterCommand(ygor.Command{
		Name:            "unalias",
		PrivMsgFunction: UnAliasCmdFunc,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	ygor.RegisterCommand(ygor.Command{
		Name:            "aliases",
		PrivMsgFunction: AliasesCmdFunc,
		Addressed:       true,
		AllowPrivate:    true,
		AllowChannel:    true,
	})
}
