// Copyright 2014, Truveris Inc. All Rights Reserved.

package main

import (
	"os/exec"
)


// Send a message to xombrero via its unix socket.
func Reboot() {
	cmd := exec.Command("sudo", "reboot")
	cmd.Start()
}
