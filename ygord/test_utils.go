// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"time"
)

var (
	fakeNow = time.Date(1982, 10, 20, 19, 0, 0, 0, time.UTC)
)

// CreateTestServer creates an ygor server for testing.
func CreateTestServer() *Server {
	srv := CreateServer(&Config{
		IRCNickname:   "whygore",
		AliasFilePath: ":memory:",
		Channels: map[string]ChannelCfg{
			"#test": ChannelCfg{},
		},
	})

	return srv
}
