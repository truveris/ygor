// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This module allows channel users to configure aliases themselves.

package main

import (
	"fmt"
	"sort"
	"strings"
)

const (
	// That should be plenty for most IRC servers to handle.
	MaxCharsPerPage = 444
)

type AliasModule struct{}

func (module AliasModule) PrivMsg(msg *PrivMsg) {}

// Command used to set a new alias.
func AliasCmdFunc(msg *PrivMsg) {
	var outputMsg string

	if len(msg.Args) == 0 {
		privMsg(msg.ReplyTo, "usage: alias name [command [params ...]]")
		return
	}

	name := msg.Args[0]
	alias := GetAlias(name)

	// Request the value of an alias.
	if len(msg.Args) == 1 {
		if alias == nil {
			privMsg(msg.ReplyTo, "error: unknown alias")
			return
		}
		privMsg(msg.ReplyTo, fmt.Sprintf("'%s' is an alias for '%s'",
			alias.Name, alias.Value))
		return
	}

	// Set a new alias.
	cmd := GetCommand(name)
	if cmd != nil {
		privMsg(msg.ReplyTo, fmt.Sprintf("error: '%s' is already a"+
			" command", name))
		return
	}

	if alias == nil {
		AddAlias(name, strings.Join(msg.Args[1:], " "))
		outputMsg = "ok (created)"
	} else {
		alias.Value = strings.Join(msg.Args[1:], " ")
		outputMsg = "ok (replaced)"
	}

	err := SaveAliases()
	if err != nil {
		outputMsg = "failed: " + err.Error()
	}

	privMsg(msg.ReplyTo, outputMsg)
}

// Take a list of a
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

// List all known aliases, by pages.
func AliasesCmdFunc(msg *PrivMsg) {
	var aliases []string

	if len(msg.Args) != 0 {
		privMsg(msg.ReplyTo, "usage: aliases")
		return
	}

	for _, alias := range Aliases {
		aliases = append(aliases, alias.Name)
	}

	sort.Strings(aliases)
	first := true
	for _, page := range getPagesOfAliases(aliases) {
		if first {
			privMsg(msg.ReplyTo, "known aliases: "+page)
			first = false
		} else {
			privMsg(msg.ReplyTo, "... "+page)
		}
	}
}

func (module AliasModule) Init() {
	RegisterCommand(Command{
		Name:         "alias",
		Function:     AliasCmdFunc,
		Addressed:    true,
		AllowDirect:  false,
		AllowChannel: true,
	})

	RegisterCommand(Command{
		Name:         "aliases",
		Function:     AliasesCmdFunc,
		Addressed:    true,
		AllowDirect:  true,
		AllowChannel: true,
	})
}
