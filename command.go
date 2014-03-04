// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

var (
	RegisteredCommands = make(map[string]Command)
)

type CommandFunction func(*PrivMsg)

type ToggleFunction func(*PrivMsg) bool

type Command struct {
	// How to call this command from IRC.
	Name string

	// If this command should be triggered by regexp.
	ToggleFunction ToggleFunction

	// Function executed when the name is called.
	Function CommandFunction

	// Define whether we expect this command to be run with the nickname as
	// prefix or without. E.g. "ygor: hello" vs just "hello".
	Addressed bool

	// Set to true if this command can be run in private (the owner can
	// always run everything in private)
	AllowDirect bool

	// Set to true of this command can be issued in a channel.
	AllowChannel bool
}

// Check if the given message matches the command.
func (cmd Command) MessageMatches(msg *PrivMsg) bool {
	// Not even the right command.
	if cmd.ToggleFunction != nil {
		if !cmd.ToggleFunction(msg) {
			return false
		}
	} else if cmd.Name != msg.Command {
		return false
	}

	// Only filter out if the command is expecting an addressed message and
	// the message isn't.
	if cmd.Addressed && !msg.Addressed {
		return false
	}

	// The owner can run any commands in private.
	if msg.IsFromOwner() {
		return true
	}

	// Some users may as well run direct commands.
	if !cmd.AllowDirect && msg.Direct {
		return false
	}

	return true
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
