// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// IRCMessageFunction is used as a type of function that receives a message from IRC
type IRCMessageFunction func(*Server, *IRCInputMessage)

// ToggleFunction is a type of function that is used to check if a module
// should be executed based on the provided Message.  It should return a
// boolean.
type ToggleFunction func(*Server, *IRCInputMessage) bool

// Command is the definition of a command to be executed either when a message
// is received from users (IRC) or from minions.
type Command struct {
	// How to call this command from IRC.
	Name string

	// If this command should be triggered by regexp.
	ToggleFunction ToggleFunction

	// Function executed when the command is called from IRC.
	PrivMsgFunction IRCMessageFunction

	// Define whether we expect this command to be run with the nickname as
	// prefix or without. E.g. "ygor: hello" vs just "hello".
	Addressed bool

	// Set to true if this command can be issued in private.
	AllowPrivate bool

	// Set to true of this command can be issued in a channel.
	AllowChannel bool
}

// IRCMessageMatches checks if the given Message matches the command.
func (cmd Command) IRCMessageMatches(srv *Server, msg *IRCInputMessage) bool {
	// Not even the right command.
	if cmd.ToggleFunction != nil {
		if !cmd.ToggleFunction(srv, msg) {
			return false
		}
	} else if cmd.Name != msg.Command {
		return false
	}

	// Check if the command forbids private messages.
	if !cmd.AllowPrivate && msg.Type == IRCInputMsgTypeIRCPrivate {
		return false
	}

	return true
}
