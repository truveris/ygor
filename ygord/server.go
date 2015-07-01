// Copyright (c) 2015 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"crypto/rand"
	"errors"
	"io"
	"log"

	"github.com/truveris/sqs"
	"github.com/truveris/ygor/ygord/alias"
)

type Server struct {
	Aliases            *alias.File
	Minions            *MinionsFile
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

	srv.Minions, err = OpenMinionsFile(config.MinionsFilePath)
	if err != nil {
		log.Fatal("minions file error: ", err.Error())
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

// GetMinionsByChannel returns all the minions configured for that channel.
func (srv *Server) GetMinionsByChannel(channel string) []*Minion {
	var minions []*Minion

	for _, name := range srv.Config.GetMinionsByChannel(channel) {
		minion, err := srv.Minions.Get(name)
		if err != nil {
			log.Printf("ignoring '%s': %s", name, err.Error())
			continue
		}
		minions = append(minions, minion)
	}

	return minions
}

// GetQueueURLsByChannel returns an array of queue URLs. These URLs are
// extracted from the minions attached to this channel.
func (srv *Server) GetQueueURLsByChannel(channel string) ([]string, error) {
	var urls []string

	minions := srv.GetMinionsByChannel(channel)

	for _, minion := range minions {
		if minion.QueueURL == "" {
			log.Printf("minion '%s' has no QueueURL", minion.Name)
			continue
		}

		urls = append(urls, minion.QueueURL)
	}

	return urls, nil
}

// StartAdapters starts all the IO adapters (IRC, Stdin/Stdout, Minions, API)
func (srv *Server) StartAdapters() (<-chan error, error) {
	cfg := srv.Config
	err := srv.StartHTTPAdapter(cfg.HTTPServerAddress)
	if err != nil {
		return nil, errors.New("error starting http adapter: " +
			err.Error())
	}

	// client, err := sqs.NewClient(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey, cfg.AWSRegionCode)
	// if err != nil {
	// 	return nil, err
	// }

	err = srv.StartIRCAdapter()
	if err != nil {
		return nil, errors.New("error starting IRC adapter: " +
			err.Error())
	}

	minionerrch := make(chan error, 0)
	// minionerrch, err := StartMinionAdapter(client, srv.Config.QueueName)
	// if err != nil {
	// 	return nil, errors.New("error starting minion adapter: " +
	// 		err.Error())
	// }

	return minionerrch, nil
}

func (srv *Server) GetSQSClient() (*sqs.Client, error) {
	return sqs.NewClient(srv.Config.AWSAccessKeyID, srv.Config.AWSSecretAccessKey,
		srv.Config.AWSRegionCode)
}

// SendToChannelMinions sends a message to all the minions of the given
// channel.
func (srv *Server) SendToChannelMinions(channel, msg string) {
	urls, err := srv.GetQueueURLsByChannel(channel)
	if err != nil {
		log.Printf("error: unable to load queue URLs, %s", err.Error())
		return
	}

	// Send the same exact data to all this channel's minion.
	for _, url := range urls {
		srv.OutputQueue <- &OutputMessage{
			Type:     OutMsgTypeMinion,
			QueueURL: url,
			Body:     sqs.SQSEncode(msg),
		}
	}

	for _, client := range srv.GetClientsByChannel(channel) {
		if client.IsAlive() {
			client.Queue <- msg
		} else {
			srv.PurgeClient(client)
		}
	}
}

// SendToQueue sends a message to our friendly minion via its SQS queue.
func (srv *Server) SendToQueue(queueURL, msg string) {
	srv.OutputQueue <- &OutputMessage{
		Type:     OutMsgTypeMinion,
		QueueURL: queueURL,
		Body:     sqs.SQSEncode(msg),
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
