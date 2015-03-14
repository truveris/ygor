// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This module allows channel users to configure aliases themselves.

package main

import (
	"errors"
	"fmt"
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

	// MaxAliasIncrements defines how far we should check for increments
	// within the same namespace.  This is really just here to avoid abuse
	// more than anything.
	MaxAliasIncrements = 10000
)

var (
	errTooManyHashes  = errors.New("too many '#'")
	errTooManyAliases = errors.New("too many aliases with this prefix")
)

// getIncrementedName returns the next available alias with a trailing number
// incremented if needed.  This is used when an alias has a trailing '#'.
func getIncrementedName(name, value string) (string, error) {
	cnt := strings.Count(name, "#")
	if cnt == 0 {
		return name, nil
	} else if cnt > 1 {
		return "", errTooManyHashes
	}

	var newName string

	for i := 1; ; i++ {
		newName = strings.Replace(name, "#", fmt.Sprintf("%d", i), 1)

		if i > MaxAliasIncrements {
			return "", errTooManyAliases
		}

		alias := Aliases.Get(newName)
		if alias == nil {
			break
		}

		if alias.Value == value {
			return "", errors.New("already exists as '" + alias.Name + "'")
		}
	}

	return newName, nil
}

// AliasModule controls all the alias-related commands.
type AliasModule struct{}

// AliasPrivMsg is the message handler for user 'alias' requests.
func (module *AliasModule) AliasPrivMsg(msg *Message) {
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
		IRCPrivMsg(msg.ReplyTo, fmt.Sprintf("%s=\"%s\" (created by %s on %s)",
			alias.Name, alias.Value, alias.Author, alias.HumanTime()))
		return
	}

	// Set a new alias.
	cmd := GetCommand(name)
	if cmd != nil {
		IRCPrivMsg(msg.ReplyTo, fmt.Sprintf("error: '%s' is a"+
			" command", name))
		return
	}

	newValue := strings.Join(msg.Args[1:], " ")

	if alias == nil {
		var creationTime time.Time

		newName, err := getIncrementedName(name, newValue)
		if err != nil {
			IRCPrivMsg(msg.ReplyTo, "error: "+err.Error())
			return
		}
		if newName != name {
			outputMsg = "ok (created as \"" + newName + "\")"
		} else {
			outputMsg = "ok (created)"
		}
		if cfg.TestMode {
			creationTime = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		} else {
			creationTime = time.Now()
		}
		Aliases.Add(newName, newValue, msg.UserID, creationTime)
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
func (module *AliasModule) UnAliasPrivMsg(msg *Message) {
	if len(msg.Args) != 1 {
		IRCPrivMsg(msg.ReplyTo, "usage: unalias name")
		return
	}

	name := msg.Args[0]
	alias := Aliases.Get(name)

	if alias == nil {
		IRCPrivMsg(msg.ReplyTo, "error: unknown alias")
		return
	}

	Aliases.Delete(name)
	Aliases.Save()
	IRCPrivMsg(msg.ReplyTo, "ok (deleted)")
}

// AliasesPrivMsg is the message handler for user 'aliases' requests.  It lists
// all the available aliases.
func (module *AliasModule) AliasesPrivMsg(msg *Message) {
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

// GrepPrivMsg is the message handler for user 'grep' requests.  It lists
// all the available aliases matching the provided pattern.
func (module *AliasModule) GrepPrivMsg(msg *Message) {
	if len(msg.Args) != 1 {
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

// RandomPrivMsg is the message handler for user 'random' requests.  It picks a
// random alias to execute based on the provided pattern or no pattern at all.
func (module *AliasModule) RandomPrivMsg(msg *Message) {
	var names []string

	switch len(msg.Args) {
	case 0:
		names = Aliases.Names()
	case 1:
		names = Aliases.Find(msg.Args[0])
	default:
		IRCPrivMsg(msg.ReplyTo, "usage: random [pattern]")
		return
	}

	if len(names) <= 0 {
		IRCPrivMsg(msg.ReplyTo, "no matches found")
		return
	}

	idx := rand.Intn(len(names))

	body, err := Aliases.Resolve(names[idx])
	if err != nil {
		IRCPrivMsg(msg.ReplyTo, "failed to resolve aliases: "+
			err.Error())
		return
	}

	newmsgs, err := NewMessagesFromBody(body)
	if err != nil {
		IRCPrivMsg(msg.ReplyTo, "error: failed to expand chose alias '"+
			names[idx]+"': "+err.Error())
		return
	}

	IRCPrivMsg(msg.ReplyTo, "the codes have chosen "+names[idx])

	for _, newmsg := range newmsgs {
		newmsg.ReplyTo = msg.ReplyTo
		newmsg.Type = msg.Type
		newmsg.UserID = msg.UserID
		newmsg.ReplyTo = msg.ReplyTo
		if newmsg == nil {
			Debug("failed to convert PRIVMSG")
			return
		}
		InputQueue <- newmsg
	}
}

// Init registers all the commands for this module.
func (module *AliasModule) Init() {
	RegisterCommand(Command{
		Name:            "alias",
		PrivMsgFunction: module.AliasPrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	RegisterCommand(Command{
		Name:            "grep",
		PrivMsgFunction: module.GrepPrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	RegisterCommand(Command{
		Name:            "random",
		PrivMsgFunction: module.RandomPrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	RegisterCommand(Command{
		Name:            "unalias",
		PrivMsgFunction: module.UnAliasPrivMsg,
		Addressed:       true,
		AllowPrivate:    false,
		AllowChannel:    true,
	})

	RegisterCommand(Command{
		Name:            "aliases",
		PrivMsgFunction: module.AliasesPrivMsg,
		Addressed:       true,
		AllowPrivate:    true,
		AllowChannel:    true,
	})
}
