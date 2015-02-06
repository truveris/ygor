// Copyright 2014-2015, Truveris Inc. All Rights Reserved.

package main

import (
	"os/exec"
)

// Reboot orders a system reboot to the minion.
func Reboot() {
	cmd := exec.Command("sudo", "reboot")
	cmd.Start()
}
