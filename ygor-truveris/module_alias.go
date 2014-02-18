// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This module allows channel users to configure aliases themselves.

package main

import (
	"fmt"
	"strings"
)

type AliasModule struct{}

func (module AliasModule) PrivMsg(msg *PrivMsg) {}

func AliasCmdFunc(where string, params []string) {
	if len(params) == 0 {
		privMsg(where, "usage: alias name [command [params ...]]")
		return
	}

	name := params[0]
	alias := GetAlias(name)

	// Request the value of an alias.
	if len(params) == 1 {
		if alias == nil {
			privMsg(where, "error: unknown alias")
			return
		}
		msg := fmt.Sprintf("'%s' is an alias for '%s'", alias.Name,
			alias.Value)
		privMsg(where, msg)
		return
	}

	// Set a new alias.
	cmd := GetCommand(name)
	if cmd != nil {
		msg := fmt.Sprintf("error: '%s' is already a command", name)
		privMsg(where, msg)
		return
	}

	if alias == nil {
		AddAlias(name, strings.Join(params[1:], " "))
		privMsg(where, "ok (created)")
	} else {
		alias.Value = strings.Join(params[1:], " ")
		privMsg(where, "ok (replaced)")
	}

	SaveAliases()
}

func AliasesCmdFunc(where string, params []string) {
	var aliases []string

	if len(params) != 0 {
		privMsg(where, "usage: aliases")
		return
	}

	for _, alias := range Aliases {
		aliases = append(aliases, alias.Name)
	}

	privMsg(where, "known aliases: "+strings.Join(aliases, ", "))
}

func (module AliasModule) Init() {
	RegisterCommand(NewCommandFromFunction("alias", AliasCmdFunc))
	RegisterCommand(NewCommandFromFunction("aliases", AliasesCmdFunc))
}
