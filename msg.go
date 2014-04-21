// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package ygor

import (
	"github.com/truveris/sqs"
)

type MsgType int

const (
	MsgTypeUnknown    MsgType = iota
	MsgTypeIRCChannel MsgType = iota
	MsgTypeIRCPrivate MsgType = iota
	MsgTypeAPI        MsgType = iota
	MsgTypeMinion     MsgType = iota
)

type Message struct {
	Type       MsgType
	SQSMessage *sqs.Message
	UserID     string
	Command    string
	Body       string
	// In the case of an IRC message, this is a nickname.
	ReplyTo    string
	Args       []string
}

func NewMessage() *Message {
	msg := &Message{}
	msg.Type = MsgTypeUnknown
	return msg
}
