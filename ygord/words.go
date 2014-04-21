// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"math/rand"
)

var (
	okWords = []string{"ok", "sure", "yes", "done"}
)

// Return a positive acknowledgement word.
func okWord() string {
	var idx int
	if cfg.TestMode {
		idx = 0
	} else {
		idx = rand.Intn(len(okWords))
	}
	return okWords[idx]
}
