// Copyright (c) 2015 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"errors"
	"log"

	"github.com/thoj/go-ircevent"
	"github.com/truveris/sqs"
	"github.com/truveris/ygor/ygord/alias"
)

type Server struct {
	Aliases *alias.File
	Minions *MinionsFile
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

	return srv
}

// GetChannelsByMinionName returns a list of channels given a minion name.
func (srv *Server) GetChannelsByMinionName(name string) []string {
	var channels []string

	for channelName, channelCfg := range srv.Config.Channels {
		for _, minionName := range channelCfg.Minions {
			if minionName == name {
				channels = append(channels, channelName)
				break
			}
		}
	}

	return channels
}

// GetChannelMinions returns all the minions configured for that channel.
func (srv *Server) GetChannelMinions(channel string) []*Minion {
	channelCfg, exists := srv.Config.Channels[channel]
	if !exists {
		log.Printf("error: %s has no queue(s) configured", channel)
		return nil
	}

	minions, err := channelCfg.GetMinions(srv)
	if err != nil {
		log.Printf("error: GetChannelMinions: %s", err.Error())
	}

	return minions
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
	client, err := srv.GetSQSClient()
	if err != nil {
		log.Printf("error: %s", err.Error())
		return
	}

	channelCfg, exists := srv.Config.Channels[channel]
	if !exists {
		log.Printf("error: %s has no queue(s) configured", channel)
		return
	}

	urls, err := channelCfg.GetQueueURLs(srv)
	if err != nil {
		log.Printf("error: unable to load queue URLs, %s", err.Error())
		return
	}

	// Send the same exact data to all this channel's minion.
	for _, url := range urls {
		err := client.SendMessage(url, sqs.SQSEncode(msg))
		if err != nil {
			log.Printf("error sending to minion: %s", err.Error())
			continue
		}
	}
}

// SendToQueue sends a message to our friendly minion via its SQS queue.
func (srv *Server) SendToQueue(queueURL, msg string) error {
	client, err := srv.GetSQSClient()
	if err != nil {
		return err
	}

	err = client.SendMessage(queueURL, msg)
	if err != nil {
		log.Printf("error sending to minion: %s", err.Error())
	}

	return nil
}

// RegisterModule adds a module to our global registry.
func (srv *Server) RegisterModule(module Module) {
	module.Init()
	modules = append(modules, module)
}

func (srv *Server) StartIRCAdapter() error {
	cfg := srv.Config
	conn = irc.IRC(cfg.IRCNickname, cfg.IRCNickname)
	//conn.VerboseCallbackHandler = true
	//conn.Debug = true

	err := conn.Connect(cfg.IRCServer)
	if err != nil {
		return err
	}

	conn.AddCallback("001", func(e *irc.Event) {
		for _, c := range cfg.GetAutoJoinChannels() {
			conn.Join(c)
		}
	})

	conn.AddCallback("PRIVMSG", func(e *irc.Event) {
		msgs := srv.NewMessagesFromEvent(e)
		for _, msg := range msgs {
			srv.IRCMessageHandler(msg)
		}
	})

	return nil
}
