// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import "os"

func startBot() {
	conn, err := connect()
	if err != nil {
		logger(err.Error())
		return
	}

	// These channels will represent the lines coming and going to the IRC
	// server.
	irc_incoming := make(chan string)
	irc_outgoing := make(chan string)
	irc_disconnect := make(chan string)
	go connectionReader(conn, irc_incoming, irc_disconnect)
	go connectionWriter(conn, irc_outgoing)

	// These channels will represent the lines coming and going to the
	// speak program.
	cmd_incoming := make(chan string)
	cmd_outgoing := make(chan string)
	go spawnHandlerProcess(cmd_incoming, cmd_outgoing)

	for {
		select {
		case data := <-irc_incoming:
			// Pass the IRC data to the soul.
			cmd_incoming <- data
		case data := <-cmd_outgoing:
			// Pass the soul data to the IRC server.
			irc_outgoing <- data
		case data := <-irc_disconnect:
			// Server has disconnected, we're done.
			logger(data)
			os.Exit(0)
		}
	}
}

func main() {
	parseCommandLine()
	logger("booting %s", cfg.Nickname)
	startBot()
}
