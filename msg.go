// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package ygor

const (
	MsgTypeUnknown   = iota
	MsgTypePrivMsg   = iota
	MsgTypeMinionMsg = iota
)

// Return the message type.
func GetMsgType(line string) int {
	if rePrivMsg.FindStringSubmatch(line) != nil {
		return MsgTypePrivMsg
	}

	if reMinionMsg.FindStringSubmatch(line) != nil {
		return MsgTypeMinionMsg
	}

	return MsgTypeUnknown
}
