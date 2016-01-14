// Copyright (c) 2015-2016 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"crypto/rand"
	"io"
	"log"
	"strings"

	"github.com/truveris/ygor/ygord/alias"
)

// Server is the main internal struct representing ygord.  This struct is a
// singleton and is available in most sub-system to access the current state
// or the configuration struct.
type Server struct {
	Aliases            *alias.File
	ClientRegistry     map[string]*Client
	InputQueue         chan *InputMessage
	OutputQueue        chan *OutputMessage
	Modules            []Module
	RegisteredCommands map[string]Command
	Salt               []byte
	*Config
}

// CreateServer produces the Server singleton with all its attributes set to
// proper defaults and all the channels initialized.
func CreateServer(config *Config) *Server {
	var err error
	srv := &Server{Config: config}

	srv.Aliases, err = alias.Open(config.AliasFilePath)
	if err != nil {
		log.Fatal("alias file error: ", err.Error())
	}

	srv.RegisteredCommands = make(map[string]Command)
	srv.InputQueue = make(chan *InputMessage, 128)
	srv.OutputQueue = make(chan *OutputMessage, 128)

	srv.ClientRegistry = make(map[string]*Client)

	srv.Salt = make([]byte, 32)
	_, err = io.ReadFull(rand.Reader, srv.Salt)
	if err != nil {
		log.Fatal("failed to generate startup salt")
	}

	return srv
}

func (srv *Server) SendToClient(client *Client, cmd ClientCommand) {
	if client.IsAlive() {
		client.Queue <- cmd
	} else {
		srv.UnregisterClient(client)
	}
}

// SendToChannelMinions sends a message to all the minions of the given
// channel.
func (srv *Server) SendToChannelMinions(channel string, cmd ClientCommand) {
	// If that channel is really just a client ID, just send it there (this
	// is done by the screensaver module for example to reach a particular
	// client).
	if client, ok := srv.ClientRegistry[channel]; ok {
		srv.SendToClient(client, cmd)
		return
	}

	channel = strings.TrimPrefix(channel, "#")
	for _, client := range srv.GetClientsByChannel(channel) {
		srv.SendToClient(client, cmd)
	}
}

// RegisterModule adds a module to our global registry.
func (srv *Server) RegisterModule(module Module) {
	module.Init(srv)
	srv.Modules = append(srv.Modules, module)
}

// RegisterCommand adds a command to the registry.  There could be only one
// command registered for each name.
func (srv *Server) RegisterCommand(cmd Command) {
	srv.RegisteredCommands[cmd.Name] = cmd
}

// GetCommand returns a registered command or nil.
func (srv *Server) GetCommand(name string) *Command {
	if cmd, ok := srv.RegisteredCommands[name]; ok {
		return &cmd
	}

	return nil
}

// FlushOutputQueue removes every single messages from the
// OutputQueue and returns them in the form of an array.
func (srv *Server) FlushOutputQueue() []*OutputMessage {
	var msgs []*OutputMessage

	for {
		select {
		case msg := <-srv.OutputQueue:
			msgs = append(msgs, msg)
		default:
			goto end
		}
	}

end:
	return msgs
}

// FlushInputQueue removes every single messages from the InputQueue and
// returns them in the form of an array.
func (srv *Server) FlushInputQueue() []*InputMessage {
	var msgs []*InputMessage

	for {
		select {
		case msg := <-srv.InputQueue:
			msgs = append(msgs, msg)
		default:
			goto end
		}
	}

end:
	return msgs
}
