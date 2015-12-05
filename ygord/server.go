// Copyright (c) 2015 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"crypto/rand"
	"errors"
	"io"
	"log"

	"github.com/truveris/ygor/ygord/alias"
)

type Server struct {
	Aliases            *alias.File
	ClientRegistry     map[string]*Client
	InputQueue         chan *Message
	OutputQueue        chan *OutputMessage
	Modules            []Module
	RegisteredCommands map[string]Command
	Salt               []byte
	*Config
}

func CreateServer(config *Config) *Server {
	var err error
	srv := &Server{Config: config}

	// We have global alias and minions files available to everyone. The
	// alias module and irc io adapter use aliases and everything uses minions.
	srv.Aliases, err = alias.Open(config.AliasFilePath)
	if err != nil {
		log.Fatal("alias file error: ", err.Error())
	}

	srv.RegisteredCommands = make(map[string]Command)
	srv.InputQueue = make(chan *Message, 128)
	srv.OutputQueue = make(chan *OutputMessage, 128)

	srv.ClientRegistry = make(map[string]*Client)

	srv.Salt = make([]byte, 32)
	_, err = io.ReadFull(rand.Reader, srv.Salt)
	if err != nil {
		log.Fatal("failed to generate startup salt")
	}

	return srv
}

// StartAdapters starts all the IO adapters (IRC, Stdin/Stdout, Minions, API)
func (srv *Server) StartAdapters() (<-chan error, error) {
	cfg := srv.Config
	err := srv.StartHTTPAdapter(cfg.HTTPServerAddress)
	if err != nil {
		return nil, errors.New("error starting http adapter: " +
			err.Error())
	}

	err = srv.StartIRCAdapter()
	if err != nil {
		return nil, errors.New("error starting IRC adapter: " +
			err.Error())
	}

	minionerrch := make(chan error, 0)

	return minionerrch, nil
}

// SendToChannelMinions sends a message to all the minions of the given
// channel.
func (srv *Server) SendToChannelMinions(channel, msg string) {
	for _, client := range srv.GetClientsByChannel(channel) {
		if client.IsAlive() {
			client.Queue <- msg
		} else {
			srv.PurgeClient(client)
		}
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

// FlushOutputQueue removes every single messages from the OutputQueue and
// returns them in the form of an array.
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
func (srv *Server) FlushInputQueue() []*Message {
	var msgs []*Message

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
