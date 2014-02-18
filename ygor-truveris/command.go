// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
)

var (
	RegisteredCommands = make(map[string]Command)
)

type CommandFunction func(string, []string)

type Command struct {
	// How to call this command from IRC.
	Name string

	// Functin executed when the name is called.
	Function CommandFunction

	// Define whether we expect this command to be run with the nickname as
	// prefix or without. E.g. "ygor: hello" vs just "hello".
	IsAddressed bool
}

// Return a registered command or nil.
func GetCommand(name string) *Command {
	if cmd, ok := RegisteredCommands[name]; ok {
		return &cmd
	}

	return nil
}

// Register this command by name. There could be only one.
func RegisterCommand(cmd Command) {
	RegisteredCommands[cmd.Name] = cmd
}

func NewCommand(name string) Command {
	cmd := Command{}
	cmd.Name = name
	cmd.IsAddressed = true
	return cmd
}

func NewCommandFromFunction(name string, f CommandFunction) Command {
	cmd := NewCommand(name)
	cmd.Function = f
	return cmd
}
