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
func (module *AliasModule) AliasPrivMsg(srv *Server, msg *Message) {
	var outputMsg string

	if len(msg.Args) == 0 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: alias name [expr ...]")
		return
	}

	name := msg.Args[0]
	alias := srv.Aliases.Get(name)

	// Request the value of an alias.
	if len(msg.Args) == 1 {
		if alias == nil {
			srv.IRCPrivMsg(msg.ReplyTo, "error: unknown alias")
			return
		}
		srv.IRCPrivMsg(msg.ReplyTo, fmt.Sprintf("%s=\"%s\" (created by %s on %s)",
			alias.Name, alias.Value, alias.Author, alias.HumanTime()))
		return
	}

	// Set a new alias.
	cmd := srv.GetCommand(name)
	if cmd != nil {
		srv.IRCPrivMsg(msg.ReplyTo, fmt.Sprintf("error: '%s' is a"+
			" command", name))
		return
	}

	newValue := strings.Join(msg.Args[1:], " ")

	if alias == nil {
		var creationTime time.Time

		newName, err := srv.Aliases.GetIncrementedName(name, newValue)
		if err != nil {
			srv.IRCPrivMsg(msg.ReplyTo, "error: "+err.Error())
			return
		}
		if newName != name {
			outputMsg = "ok (created as \"" + newName + "\")"
		} else {
			outputMsg = "ok (created)"
		}
		creationTime = time.Now()
		srv.Aliases.Add(newName, newValue, msg.UserID, creationTime)
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

	srv.IRCPrivMsg(msg.ReplyTo, outputMsg)
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
func (module *AliasModule) UnAliasPrivMsg(srv *Server, msg *Message) {
	if len(msg.Args) != 1 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: unalias name")
		return
	}

	name := msg.Args[0]
	alias := srv.Aliases.Get(name)

	if alias == nil {
		srv.IRCPrivMsg(msg.ReplyTo, "error: unknown alias")
		return
	}

	srv.Aliases.Delete(name)
	srv.Aliases.Save()
	srv.IRCPrivMsg(msg.ReplyTo, "ok (deleted)")
}

// AliasesPrivMsg is the message handler for user 'aliases' requests.  It lists
// all the available aliases.
func (module *AliasModule) AliasesPrivMsg(srv *Server, msg *Message) {
	if len(msg.Args) != 0 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: aliases")
		return
	}

	aliases := srv.Aliases.Names()

	if len(aliases) > MaxAliasesForFullList {
		srv.IRCPrivMsg(msg.ReplyTo, "error: too many results, use grep")
		return
	}

	sort.Strings(aliases)
	first := true
	for _, page := range getPagesOfAliases(aliases) {
		if first {
			srv.IRCPrivMsg(msg.ReplyTo, "known aliases: "+page)
			first = false
		} else {
			srv.IRCPrivMsg(msg.ReplyTo, "... "+page)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// GrepPrivMsg is the message handler for user 'grep' requests.  It lists
// all the available aliases matching the provided pattern.
func (module *AliasModule) GrepPrivMsg(srv *Server, msg *Message) {
	if len(msg.Args) != 1 {
		srv.IRCPrivMsg(msg.ReplyTo, "usage: grep pattern")
		return
	}

	results := srv.Aliases.Find(msg.Args[0])
	sort.Strings(results)

	if len(results) == 0 {
		srv.IRCPrivMsg(msg.ReplyTo, "error: no matches found")
		return
	}

	found := strings.Join(results, ", ")
	if len(found) > MaxCharsPerPage {
		srv.IRCPrivMsg(msg.ReplyTo, "error: too many matches, refine your search")
		return
	}

	srv.IRCPrivMsg(msg.ReplyTo, found)
}

// RandomPrivMsg is the message handler for user 'random' requests.  It picks a
// random alias to execute based on the provided pattern or no pattern at all.
func (module *AliasModule) RandomPrivMsg(srv *Server, msg *Message) {
	var names []string

	switch len(msg.Args) {
	case 0:
		names = srv.Aliases.Names()
	case 1:
		names = srv.Aliases.Find(msg.Args[0])
	default:
		srv.IRCPrivMsg(msg.ReplyTo, "usage: random [pattern]")
		return
	}

	if len(names) <= 0 {
		srv.IRCPrivMsg(msg.ReplyTo, "error: no matches found")
		return
	}

	idx := rand.Intn(len(names))

	body, err := srv.Aliases.Resolve(names[idx], 0)
	if err != nil {
		srv.IRCPrivMsg(msg.ReplyTo, "error: failed to resolve aliases: "+
			err.Error())
		return
	}

	newmsgs, err := srv.NewMessagesFromBody(body)
	if err != nil {
		srv.IRCPrivMsg(msg.ReplyTo, "error: failed to expand chosen alias '"+
			names[idx]+"': "+err.Error())
		return
	}

	srv.IRCPrivAction(msg.ReplyTo, "chooses "+names[idx])

	for _, newmsg := range newmsgs {
		newmsg.ReplyTo = msg.ReplyTo
		newmsg.Type = msg.Type
		newmsg.UserID = msg.UserID
		if newmsg == nil {
			log.Printf("failed to convert PRIVMSG")
			return
		}
		srv.InputQueue <- newmsg
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
