// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// IRCInputMsgType is used to categorize the message type constants below.
type IRCInputMsgType int

// IRCOutMsgType is used to categorize the outgoing message type constants below.
type IRCOutMsgType int

// Types of messages received from any various source (IRC, minions, etc.).
// The first types are used for communication between the different components.
// The Exit and Fatal types are used for flow control and are mostly triggered
// internally (e.g. IO error).
const (
	IRCInputMsgTypeUnknown    IRCInputMsgType = iota
	IRCInputMsgTypeIRCChannel IRCInputMsgType = iota
	IRCInputMsgTypeIRCPrivate IRCInputMsgType = iota

	IRCOutMsgTypePrivMsg IRCOutMsgType = iota
	IRCOutMsgTypeAction  IRCOutMsgType = iota
)

// IRCOutputMessage is the representation of an outbound IRC message.
type IRCOutputMessage struct {
	Type    IRCOutMsgType
	Channel string
	Body    string
}

// IRCInputMessage is a representation of an incoming IRC message
type IRCInputMessage struct {
	Type     IRCInputMsgType
	Nickname string
	Command  string
	Body     string
	// ReplyTo could be a nickname or a channel.
	ReplyTo string
	Args    []string

	// Recursion tracks the recursion level in case commands create/call
	// other commands and produce more messages.  A message create out of
	// the IRC handler will have 0 recursion but modules generating more
	// messages from it should increment it.
	Recursion int
}

// NewIRCInputMessage allocates a new message without type.
func NewIRCInputMessage() *IRCInputMessage {
	msg := &IRCInputMessage{}
	msg.Type = IRCInputMsgTypeUnknown
	return msg
}
