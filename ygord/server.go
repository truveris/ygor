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

// Server is the main internal struct representing ygord.  This struct is a
// singleton and is available in most sub-system to access the current state
// or the configuration struct.
type Server struct {
	Aliases            *alias.File
	ClientRegistry     map[string]*Client
	IRCInputQueue      chan *IRCInputMessage
	IRCOutputQueue     chan *IRCOutputMessage
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

	// We have global alias and minions files available to everyone. The
	// alias module and irc io adapter use aliases and everything uses minions.
	srv.Aliases, err = alias.Open(config.AliasFilePath)
	if err != nil {
		log.Fatal("alias file error: ", err.Error())
	}

	srv.RegisteredCommands = make(map[string]Command)
	srv.IRCInputQueue = make(chan *IRCInputMessage, 128)
	srv.IRCOutputQueue = make(chan *IRCOutputMessage, 128)

	srv.ClientRegistry = make(map[string]*Client)

	srv.Salt = make([]byte, 32)
	_, err = io.ReadFull(rand.Reader, srv.Salt)
	if err != nil {
		log.Fatal("failed to generate startup salt")
	}

	return srv
}

// StartAdapters starts all the IO adapters (IRC, Stdin/Stdout, Minions, API)
func (srv *Server) StartAdapters() error {
	cfg := srv.Config
	err := srv.StartHTTPAdapter(cfg.HTTPServerAddress)
	if err != nil {
		return errors.New("error starting http adapter: " +
			err.Error())
	}

	err = srv.StartIRCAdapter()
	if err != nil {
		return errors.New("error starting IRC adapter: " +
			err.Error())
	}

	return nil
}

// SendToChannelMinions sends a message to all the minions of the given
// channel.
func (srv *Server) SendToChannelMinions(channel, msg string) {
	for _, client := range srv.GetClientsByChannel(channel) {
		if client.IsAlive() {
			client.Queue <- msg
		} else {
			srv.UnregisterClient(client)
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

// FlushIRCOutputQueue removes every single messages from the
// IRCOutputQueue and returns them in the form of an array.
func (srv *Server) FlushIRCOutputQueue() []*IRCOutputMessage {
	var msgs []*IRCOutputMessage

	for {
		select {
		case msg := <-srv.IRCOutputQueue:
			msgs = append(msgs, msg)
		default:
			goto end
		}
	}

end:
	return msgs
}

// FlushIRCInputQueue removes every single messages from the InputQueue and
// returns them in the form of an array.
func (srv *Server) FlushIRCInputQueue() []*IRCInputMessage {
	var msgs []*IRCInputMessage

	for {
		select {
		case msg := <-srv.IRCInputQueue:
			msgs = append(msgs, msg)
		default:
			goto end
		}
	}

end:
	return msgs
}
