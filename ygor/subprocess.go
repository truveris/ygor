// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"io"
	"bufio"
	"os/exec"
	"strings"
	"time"
)

// Go routine feeding the outgoing channel all the lines from
// the soul program.
func handleSoulOutput(bufReader *bufio.Reader, outgoing, dead_process chan string) {
	for {
		output, err := bufReader.ReadString('\n')
		if err != nil {
			dead_process <- "read fail: " + err.Error()
			return
		}
		output = strings.TrimSpace(output)
		logger("[soul] > %s", output)
		outgoing <- output
	}
}

// Go routine feeding the soul with input from the IRC server.
func handleSoulInput(writer io.WriteCloser, incoming, dead_process chan string) {
	for {
		select {
		case data := <-incoming:
			data = strings.TrimSpace(data)
			logger("[soul] < %s", data)
			writer.Write([]byte(data))
			writer.Write([]byte{'\n'})
		case data := <-dead_process:
			data = strings.TrimSpace(data)
			logger("[soul] %s", data)
			return
		}
	}
}

// Build the command line arguments that will be passed to the soul.
func buildCommandLineArguments() []string {
	args := make([]string, 0)
	if len(soulArgs) > 1 {
		args = append(args, soulArgs[1:]...)
	}
	args = append(args, "--nickname=" + cfg.Nickname)
	return args
}

// Infinitely respawn a soul process until the body seems to have passed away
// (pipes are cut off).
func spawnHandlerProcess(incoming chan string, outgoing chan string) error {
	for {
		cmdArgs := buildCommandLineArguments()
		logger("[body] (re-)spawning soul: %s", cmdArgs)

		time.Sleep(1 * time.Second)
		dead_process := make(chan string)

		cmd := exec.Command(soulArgs[0], cmdArgs...)
		writer, err := cmd.StdinPipe()
		if err != nil {
			logger("[body] soul.StdinPipe: %s", err.Error())
			continue
		}
		reader, err := cmd.StdoutPipe()
		if err != nil {
			logger("[body] soul.StdoutPipe: %s", err.Error())
			continue
		}
		bufReader := bufio.NewReader(reader)

		err = cmd.Start()
		if err != nil {
			logger("[body] soul.Start: %s", err.Error())
			continue
		}

		go handleSoulOutput(bufReader, outgoing, dead_process)
		go handleSoulInput(writer, incoming, dead_process)

		cmd.Wait()
	}

	return nil
}
