// Copyright 2014, Truveris Inc. All Rights Reserved.

package main

import (
	"fmt"
	"log"
	"net"
)

// Send a message to xombrero via its unix socket.
func Xombrero(data string) {
	if cfg.TestMode {
		log.Printf("xombrero: %s", data)
		return
	}

	conn, err := net.Dial("unix", cfg.XombreroSocket)
	if err != nil {
		SendToSoul("xombrero error " + err.Error())
		log.Printf("xombrero: unable to connect to %s", err.Error())
		return
	}
	fmt.Fprintf(conn, "%s\x00", data)
	conn.Close()
	SendToSoul("xombrero ok")
}
