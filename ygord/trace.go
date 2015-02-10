// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

// waitForTraceRequest is useful as a go routine. It watches for USR1 signal
// and dumps all the stack traces in the logs.
func waitForTraceRequest() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR1)

	for _ = range ch {
		log.Printf("Received USR1 signal, printing stack trace:")
		buf := make([]byte, 40960)
		i := runtime.Stack(buf, true)
		log.Printf("%s", buf[:i])
	}
}
