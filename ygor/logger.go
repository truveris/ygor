// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"time"
)

// Basic logger printing everything on stdout with a timestamp.
func logger(format string, msgs ...interface{}) {
	now := time.Now().Format(time.RFC3339)
	fmt.Printf("[%s] %s\n", now, fmt.Sprintf(format, msgs...))
}
