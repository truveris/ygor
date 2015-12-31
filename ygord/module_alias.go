// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This module allows channel users to configure aliases themselves.

package main

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"
)

const (
	// MaxCharsPerPage is used to define the size of a page when breaking
	// results into multiple messages.
	MaxCharsPerPage = 444

	// MaxAliasesForFullList is the number of alias 'aliases' can list.
	MaxAliasesForFullList = 40
)

// AliasModule controls all the alias-related commands.
type AliasModule struct{}

// AliasPrivMsg is the message handler for user 'alias' requests.
func (module *AliasModule) AliasPrivMsg(srv *Server, msg *InputMessage) {
	var outputMsg string

	if len(msg.Args) == 0 {
		srv.Reply(msg, "usage: alias name [expr ...]")
		return
	}

	name := msg.Args[0]
	alias := srv.Aliases.Get(name)

	// Request the value of an alias.
	if len(msg.Args) == 1 {
		if alias == nil {
			srv.Reply(msg, "error: unknown alias")
			return
		}
		srv.Reply(msg, fmt.Sprintf("%s=\"%s\" (created by %s on %s)",
			alias.Name, alias.Value, alias.Author, alias.HumanTime()))
		return
	}

	// Set a new alias.
	cmd := srv.GetCommand(name)
	if cmd != nil {
		srv.Reply(msg, fmt.Sprintf("error: '%s' is a"+
			" command", name))
		return
	}

	newValue := strings.Join(msg.Args[1:], " ")

	if alias == nil {
		var creationTime time.Time

		newName, err := srv.Aliases.GetIncrementedName(name, newValue)
		if err != nil {
			srv.Reply(msg, "error: "+err.Error())
			return
		}
		if newName != name {
			outputMsg = "ok (created as \"" + newName + "\")"
		} else {
			outputMsg = "ok (created)"
		}
		creationTime = time.Now()
		srv.Aliases.Add(newName, newValue, msg.Nickname, creationTime)
	} else if alias.Value == newValue {
		outputMsg = "no changes"
	} else {
		outputMsg = "ok (replaces \"" + alias.Value + "\")"
		alias.Value = newValue
	}

	err := srv.Aliases.Save()
	if err != nil {
		outputMsg = "error: " + err.Error()
	}

	srv.Reply(msg, outputMsg)
}

// getPagesOfAliases takes a list of aliases and returns joined pages of them.
func getPagesOfAliases(aliases []string) []string {
	var pages []string
	length := 0

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

// UnAliasPrivMsg is the message handler for user 'unalias' requests.
func (module *AliasModule) UnAliasPrivMsg(srv *Server, msg *InputMessage) {
	if len(msg.Args) != 1 {
		srv.Reply(msg, "usage: unalias name")
		return
	}

	name := msg.Args[0]
	alias := srv.Aliases.Get(name)

	if alias == nil {
		srv.Reply(msg, "error: unknown alias")
		return
	}

	srv.Aliases.Delete(name)
	srv.Aliases.Save()
	srv.Reply(msg, "ok (deleted)")
}

// AliasesPrivMsg is the message handler for user 'aliases' requests.  It lists
// all the available aliases.
func (module *AliasModule) AliasesPrivMsg(srv *Server, msg *InputMessage) {
	if len(msg.Args) != 0 {
		srv.Reply(msg, "usage: aliases")
		return
	}

	aliases := srv.Aliases.Names()

	if len(aliases) > MaxAliasesForFullList {
		srv.Reply(msg, "error: too many results, use grep")
		return
	}

	sort.Strings(aliases)
	first := true
	for _, page := range getPagesOfAliases(aliases) {
		if first {
			srv.Reply(msg, "known aliases: "+page)
			first = false
		} else {
			srv.Reply(msg, "... "+page)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// GrepPrivMsg is the message handler for user 'grep' requests.  It lists
// all the available aliases matching the provided pattern.
func (module *AliasModule) GrepPrivMsg(srv *Server, msg *InputMessage) {
	if len(msg.Args) != 1 {
		srv.Reply(msg, "usage: grep pattern")
		return
	}

	results := srv.Aliases.Find(msg.Args[0])
	sort.Strings(results)

	if len(results) == 0 {
		srv.Reply(msg, "error: no matches found")
		return
	}

	found := strings.Join(results, ", ")
	if len(found) > MaxCharsPerPage {
		srv.Reply(msg, "error: too many matches, refine your search")
		return
	}

	srv.Reply(msg, found)
}

// RandomPrivMsg is the message handler for user 'random' requests.  It picks a
// random alias to execute based on the provided pattern or no pattern at all.
func (module *AliasModule) RandomPrivMsg(srv *Server, msg *InputMessage) {
	var names []string

	switch len(msg.Args) {
	case 0:
		names = srv.Aliases.Names()
	case 1:
		names = srv.Aliases.Find(msg.Args[0])
	default:
		srv.Reply(msg, "usage: random [pattern]")
		return
	}

	if len(names) <= 0 {
		srv.Reply(msg, "error: no matches found")
		return
	}

	idx := rand.Intn(len(names))

	body, err := srv.Aliases.Resolve(names[idx], 0)
	if err != nil {
		srv.Reply(msg, "error: failed to resolve aliases: "+
			err.Error())
		return
	}

	newmsgs, err := srv.NewMessagesFromBody(body, msg.Recursion+1)
	if err != nil {
		srv.Reply(msg, "error: failed to expand chosen alias '"+
			names[idx]+"': "+err.Error())
		return
	}

	srv.Reply(msg, "/me chooses "+names[idx])

	for _, newmsg := range newmsgs {
		newmsg.ReplyTo = msg.ReplyTo
		newmsg.Type = msg.Type
		newmsg.Nickname = msg.Nickname
		if newmsg == nil {
			log.Printf("failed to convert PRIVMSG")
			return
		}
		srv.IRCMessageHandler(newmsg)
	}
}

// Init registers all the commands for this module.
func (module *AliasModule) Init(srv *Server) {
	srv.RegisterCommand(Command{
		Name:            "alias",
		PrivMsgFunction: module.AliasPrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	srv.RegisterCommand(Command{
		Name:            "grep",
		PrivMsgFunction: module.GrepPrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	srv.RegisterCommand(Command{
		Name:            "random",
		PrivMsgFunction: module.RandomPrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	srv.RegisterCommand(Command{
		Name:            "unalias",
		PrivMsgFunction: module.UnAliasPrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	srv.RegisterCommand(Command{
		Name:            "aliases",
		PrivMsgFunction: module.AliasesPrivMsg,
		Addressed:       true,
		AllowPrivate:    true,
		AllowChannel:    true,
	})
}
