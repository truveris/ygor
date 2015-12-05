// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// MsgType is used to categorize the message type constants below.
type MsgType int

// OutMsgType is used to categorize the outgoing message type constants below.
type OutMsgType int

// Types of messages received from any various source (IRC, minions, etc.).
// The first types are used for communication between the different components.
// The Exit and Fatal types are used for flow control and are mostly triggered
// internally (e.g. IO error).
const (
	MsgTypeUnknown    MsgType = iota
	MsgTypeIRCChannel MsgType = iota
	MsgTypeIRCPrivate MsgType = iota
	MsgTypeAPI        MsgType = iota
	MsgTypeMinion     MsgType = iota
	MsgTypeExit       MsgType = iota
	MsgTypeFatal      MsgType = iota

	OutMsgTypePrivMsg OutMsgType = iota
	OutMsgTypeAction  OutMsgType = iota
	OutMsgTypeMinion  OutMsgType = iota
)

// OutgoingMessage is a representation of a message passed through ygord, be it IRC,
// minion, etc.
type OutputMessage struct {
	Type     OutMsgType
	Channel  string
	QueueURL string
	Body     string
}

// Message is a representation of a message passed through ygord, be it IRC,
// minion, etc.
type Message struct {
	Type    MsgType
	UserID  string
	Command string
	Body    string
	// In the case of an IRC message, this is a nickname or a channel.
	ReplyTo string
	Args    []string

	// Recursion tracks the recursion level in case commands create/call
	// other commands and produce more messages.  A message create out of
	// the IRC handler will have 0 recursion but modules generating more
	// messages from it should increment it.
	Recursion int
}

// NewMessage allocates a new message without type.
func NewMessage() *Message {
	msg := &Message{}
	msg.Type = MsgTypeUnknown
	return msg
}

// NewExitMessage allocates a new message of type Exit.
func NewExitMessage(body string) *Message {
	msg := NewMessage()
	msg.Type = MsgTypeExit
	msg.Body = body
	return msg
}

// NewFatalMessage allocates a new message of type Fatal.
func NewFatalMessage(body string) *Message {
	msg := NewMessage()
	msg.Type = MsgTypeFatal
	msg.Body = body
	return msg
}
