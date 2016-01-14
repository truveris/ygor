// Copyright 2014-2016, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"regexp"
	"strings"
)

var (
	reAddressed = regexp.MustCompile(`^\s*@?(\w+)[:,.]*\s*(.*)`)
)

// InputMsgType is used to categorize the message type constants below.
type InputMsgType int

// OutputMsgType is used to categorize the outgoing message type constants below.
type OutputMsgType int

// Types of messages to/from any various source (IRC, Mattermost, etc.).  The
// first constants (Input*) are used to represent a message ingested from the
// chat system, the second (Output*) represent a message traveling out of ygor
// to the chat system.
const (
	InputMsgTypeUnknown     InputMsgType = iota
	InputMsgTypeIRCChannel  InputMsgType = iota
	InputMsgTypeIRCPrivate  InputMsgType = iota
	InputMsgTypeMattermost  InputMsgType = iota
	InputMsgTypeScreensaver InputMsgType = iota

	OutputMsgTypePrivMsg     OutputMsgType = iota
	OutputMsgTypeAction      OutputMsgType = iota
	OutputMsgTypeMattermost  OutputMsgType = iota
	OutputMsgTypeScreensaver OutputMsgType = iota
)

// OutputMessage is the representation of an outbound IRC message.
type OutputMessage struct {
	Type    OutputMsgType
	Channel string
	Body    string
}

// InputMessage is a representation of an incoming IRC message
type InputMessage struct {
	Type     InputMsgType
	Nickname string
	Command  string
	Body     string
	// ReplyTo could be a nickname or a channel.
	ReplyTo string
	Args    []string

	// Depth tracks the recursion and depth level in case commands
	// create/call other commands and produce more messages.  A message
	// create out of the IRC handler will have 0 recursion but modules
	// generating more messages from it should increment it.
	Depth int
}

// NewInputMessage allocates a new message without type.
func NewInputMessage() *InputMessage {
	msg := &InputMessage{}
	msg.Type = InputMsgTypeUnknown
	return msg
}

func (msg *InputMessage) NewResponse(text string) *OutputMessage {
	var outputType OutputMsgType

	replyTo := msg.ReplyTo

	switch msg.Type {
	case InputMsgTypeUnknown:
		return nil
	case InputMsgTypeIRCChannel:
		if !strings.HasPrefix(replyTo, "#") {
			replyTo = "#" + replyTo
		}
		if strings.HasPrefix(text, "/me ") {
			outputType = OutputMsgTypeAction
			text = strings.TrimPrefix(text, "/me ")
		} else {
			outputType = OutputMsgTypePrivMsg
		}
	case InputMsgTypeIRCPrivate:
		if strings.HasPrefix(text, "/me ") {
			outputType = OutputMsgTypeAction
			text = strings.TrimPrefix(text, "/me ")
		} else {
			outputType = OutputMsgTypePrivMsg
		}
	case InputMsgTypeMattermost:
		outputType = OutputMsgTypeMattermost
		if strings.HasPrefix(text, "/me ") {
			text = "*" + strings.TrimPrefix(text, "/me ") + "*"
		}
	case InputMsgTypeScreensaver:
		outputType = OutputMsgTypeScreensaver
	}

	return &OutputMessage{
		Type:    outputType,
		Channel: replyTo,
		Body:    text,
	}
}

func (msg *InputMessage) IsMattermost() bool {
	if msg.Type == InputMsgTypeMattermost {
		return true
	}
	return false
}

// Reply sends a reply based on the given message (same channel, mirror messag)
func (srv *Server) Reply(msg *InputMessage, text string) {
	lines := strings.Split(text, "\n")
	for i := 0; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}
		srv.OutputQueue <- msg.NewResponse(lines[i])
	}
}
