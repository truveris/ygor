// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package ygor

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func WaitForTraceRequest() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGUSR1)

	for _ = range ch {
		log.Printf("Received USR1 signal, printing stack trace:")
		buf := make([]byte, 4096)
		runtime.Stack(buf, true)
		log.Printf("%s", buf)
	}
}
