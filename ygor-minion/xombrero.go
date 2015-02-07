// Copyright 2014-2015, Truveris Inc. All Rights Reserved.

package main

import (
	"fmt"
	"log"
	"net"
)

// Xombrero sends a message to xombrero via its unix socket.
func Xombrero(data string) {
	if cfg.TestMode {
		log.Printf("xombrero: %s", data)
		return
	}

	if cfg.XombreroSocket == "" {
		log.Printf("xombrero: no socket configured")
		return
	}

	conn, err := net.Dial("unix", cfg.XombreroSocket)
	if err != nil {
		Send("xombrero error " + err.Error())
		log.Printf("xombrero: unable to connect to %s", err.Error())
		return
	}
	fmt.Fprintf(conn, "%s\x00", data)
	conn.Close()
	Send("xombrero ok")
}
