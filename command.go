// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package ygor

var (
	RegisteredCommands = make(map[string]Command)
)

type MessageFunction func(*Message)
type ToggleFunction func(*Message) bool

type Command struct {
	// How to call this command from IRC.
	Name string

	// If this command should be triggered by regexp.
	ToggleFunction ToggleFunction

	// Function executed when the command is called from IRC.
	PrivMsgFunction MessageFunction

	// Function executed when the command is called from a minion.
	MinionMsgFunction MessageFunction

	// Define whether we expect this command to be run with the nickname as
	// prefix or without. E.g. "ygor: hello" vs just "hello".
	Addressed bool

	// Set to true if this command can be issued in private.
	AllowPrivate bool

	// Set to true of this command can be issued in a channel.
	AllowChannel bool
}

// Check if the given PrivMsg matches the command.
func (cmd Command) IRCMessageMatches(msg *Message) bool {
	// Not even the right command.
	if cmd.ToggleFunction != nil {
		if !cmd.ToggleFunction(msg) {
			return false
		}
	} else if cmd.Name != msg.Command {
		return false
	}

	// Check if the command forbids private messages.
	if !cmd.AllowPrivate && msg.Type == MsgTypeIRCPrivate {
		return false
	}

	return true
}

// Check if the given MinionMsg matches the command. We do not bother with
// ToggleFunction in this case. There is no reason to be broad since machines
// are producing the messages.
func (cmd Command) MinionMessageMatches(msg *Message) bool {
	if cmd.Name != msg.Command {
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
