// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"strings"
	"testing"
)

func AssertStringEquals(t *testing.T, a, b string) {
	if a != b {
		t.Error(fmt.Sprintf("Strings do not match (a=%s, b=%s)", a, b))
	}
}

func AssertStringHasPrefix(t *testing.T, a, b string) {
	if !strings.HasPrefix(a, b) {
		t.Error(fmt.Sprintf("String a=%s is not prefixed by b=%s", a, b))
	}
}

func AssertStringContains(t *testing.T, s, substr string) {
	if !strings.Contains(s, substr) {
		t.Error(fmt.Sprintf("String '%s' does not contain '%s'", s, substr))
	}
}

func AssertIntEquals(t *testing.T, a, b int) {
	if a != b {
		t.Error(fmt.Sprintf("Integers to do not match (a=%d, b=%d)", a, b))
	}
}

func CreateTestServerWithTwoMinions(t *testing.T) *Server {
	srv := CreateServer(&Config{
		IRCNickname:     "whygore",
		AliasFilePath:   ":memory:",
		MinionsFilePath: ":memory:",
		Channels: map[string]ChannelCfg{
			"#test": ChannelCfg{
				Minions: []string{"pi1", "pi2"},
			},
		},
	})

	return srv
}
