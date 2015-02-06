// Copyright 2015, Truveris Inc. All Rights Reserved.

package main

import (
	"log"
	"strconv"
	"time"

	"github.com/truveris/goturret"
	"github.com/truveris/gousb/usb"
)

var (
	turrets []*turret.Turret
	ctx     *usb.Context
)

func getInt(s string) int {
	if s == "" {
		return 0
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		Send("turret error parsing int")
		log.Printf("turret: error parsing int: %s", s)
		return 0
	}

	return i
}

func getShots(s string) int {
	shots := getInt(s)

	if shots < 1 {
		return 1
	}

	if shots > 4 {
		shots = 4
	}

	return shots
}

func getBlinks(s string) int {
	times := getInt(s)

	if times < 1 {
		return 1
	}

	if times > 16 {
		times = 16
	}

	return times
}

func getDuration(s string) time.Duration {
	ms, err := time.ParseDuration(s + "ms")
	if err != nil {
		Send("turret error parsing duration")
		log.Printf("turret: error parsing duration: %s", s)
	}
	return ms
}

func getBoolean(s string) bool {
	if s == "on" {
		return true
	}
	return false
}

// OpenTurrets creates a new USB context and opens all the devices known to be
// turrets.
func OpenTurrets() {
	var err error

	ctx = usb.NewContext()

	turrets, err = turret.Find(ctx)
	if err != nil {
		log.Printf("turret: %s", err)
		return
	}
}

// CloseTurrets closes all the known Turret devices and closes the USB context.
func CloseTurrets() {
	for _, t := range turrets {
		t.Close()
	}

	ctx.Close()
}

// Turret executes the turret command on the minion.
func Turret(data string) {
	cmd, value := SplitTwo(data)

	if cmd == "" {
		Send("turret error no command")
		log.Printf("turret: no command")
		return
	}

	for _, t := range turrets {
		switch cmd {
		case "blinkon":
			t.BlinkOn(getBlinks(value))
		case "blinkoff":
			t.BlinkOff(getBlinks(value))
		case "light":
			t.Light(getBoolean(value))
		case "stop":
			t.Stop()
		case "fire":
			t.Fire(getShots(value))
		case "left":
			t.Left(getDuration(value))
			t.Stop()
		case "right":
			t.Right(getDuration(value))
			t.Stop()
		case "up":
			t.Up(getDuration(value))
			t.Stop()
		case "down":
			t.Down(getDuration(value))
			t.Stop()
		case "reset":
			t.Reset()
		case "type":
			Send("turret type " + t.HumanReadableType())
		}
	}

	Send("turret ok")
}
