// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	// Connection states
	CS_INIT              = iota
	CS_WAITING_FOR_HELLO = iota
	CS_LIVE              = iota

	// Connection error codes (RFC 1459)
	ERR_NONICKNAMEGIVEN  = 431
	ERR_ERRONEUSNICKNAME = 432
	ERR_NICKNAMEINUSE    = 433
	ERR_NICKCOLLISION    = 436
)

type ServerMessage struct {
	Code     int16
	Nickname string
	Message  string
}

var (
	ConnectionState = CS_INIT
	ReServerMessage = regexp.MustCompile(`^:[^ ]+ ([0-9]{2,4}) ([^ ]+) (.*)`)
)

// Send a command to the IRC server.
func sendLine(conn net.Conn, cmd string) {
	cmd = strings.TrimSpace(cmd)
	logger("[body] > %s", cmd)
	fmt.Fprintf(conn, "%s\r\n", cmd)
}

func parseServerMessageCode(line string) int16 {
	tokens := ReServerMessage.FindStringSubmatch(line)
	if tokens == nil {
		return 0
	}

	code, err := strconv.ParseInt(tokens[1], 10, 16)
	if err != nil {
		logger("[body] invalid server message: bad code (%s) in: %s",
			err.Error(), line)
		return 0
	}

	if tokens[2] != cfg.Nickname {
		logger("[body] invalid server message: wrong nickname in: %s",
			line)
		return 0
	}

	return int16(code)
}

// Connect to the selected server and join all the specified channels.
func connect() (net.Conn, error) {
	conn, err := net.Dial("tcp", cfg.Hostname)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func connectionReader(conn net.Conn, incoming chan string, disconnect chan string) {
	bufReader := bufio.NewReader(conn)

	for {
		data, err := bufReader.ReadString('\n')
		if err == io.EOF {
			disconnect <- "server disconnected"
			break
		}
		if err != nil {
			panic(err)
		}

		switch ConnectionState {
		case CS_WAITING_FOR_HELLO:
			// This is the NICK/USER phase, add more underscores to
			// the nick, until we find one available. If we get any
			// message other than a NICK error, we assume the
			// server likes us, we move on to channel joining
			// state.
			code := parseServerMessageCode(data)

			switch code {
			case ERR_NONICKNAMEGIVEN, ERR_ERRONEUSNICKNAME,
				ERR_NICKNAMEINUSE, ERR_NICKCOLLISION:
				cfg.Nickname = cfg.Nickname + "_"
				ConnectionState = CS_INIT
			default:
				ConnectionState = CS_LIVE
				incoming <- data
			}

		case CS_LIVE:
			// Handle PING request from the server. Without these
			// our bot would time out. Don't push that to the
			// channel.
			if strings.Index(data, "PING :") == 0 {
				r := strings.Replace(data, "PING", "PONG", 1)
				fmt.Fprintf(conn, "%s", r)
				continue
			}

			incoming <- data
		// Standby
		default:
			time.Sleep(20 * time.Millisecond)
		}
	}
}

func connectionWriter(conn net.Conn, outgoing chan string) {
	for {
		switch ConnectionState {
		case CS_INIT:
			sendLine(conn, fmt.Sprintf("NICK %s", cfg.Nickname))
			sendLine(conn, fmt.Sprintf("USER %s localhost "+
				"127.0.0.1 :%s\r\n", cfg.Nickname,
				cfg.Nickname))
			ConnectionState = CS_WAITING_FOR_HELLO
		case CS_LIVE:
			for msg := range outgoing {
				fmt.Fprintf(conn, "%s\r\n", msg)
			}
		// Standby
		default:
			time.Sleep(20 * time.Millisecond)
		}
	}
}
