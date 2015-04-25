// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

// Cmd is a singleton used to store the command-line parameters.
type CmdLine struct {
	ConfigFile string `short:"c" description:"Configuration file" default:"/etc/ygord.conf"`
}

// ChannelCfg represents a per-channel grouping of minions.
type ChannelCfg struct {
	Minions []string
}

// Config is a singleton used to store the file configuration.
type Config struct {
	// AWS Region to use for SQS access (e.g. us-east-1).
	AWSRegionCode string

	// AWS access key id
	AWSAccessKeyID string
	// AWS secret access key
	AWSSecretAccessKey string

	// All the configured channels. ygord will JOIN every single one of
	// them and will push commands to the configured associated minions.
	Channels map[string]ChannelCfg

	// Queue used by ygord to receive feedback from the minions.
	QueueName string

	// Hostname and port to use to connect to the IRC server.
	IRCServer string

	// Nickname of the bot. FIXME: this is not currently synchronized
	IRCNickname string

	// Try to send debug information to this channel in lieu of log file.
	AdminChannel string

	// Any chatter from these nicks will be dropped (other bots).
	Ignore []string

	// Where to find the alias file. Will use the local alias file found in
	// the current directory by default.
	AliasFilePath string

	// Where to find the minions file.
	MinionsFilePath string

	// If defined, start a web server to list the aliases (e.g. :8989)
	HTTPServerAddress string
}

// GetAutoJoinChannels returns a list of all the auto-join channels (all unique
// configured channels and debug channels).
func (cfg *Config) GetAutoJoinChannels() []string {
	channels := make(StringSet, 0)

	for name := range cfg.Channels {
		channels.Add(name)
	}

	if cfg.AdminChannel != "" {
		channels.Add(cfg.AdminChannel)
	}

	return channels.Array()
}

// GetMinions returns an array of minions configured for this ChannelCfg.
func (channelCfg *ChannelCfg) GetMinions(srv *Server) ([]*Minion, error) {
	var minions []*Minion

	for _, name := range channelCfg.Minions {
		minion, err := srv.Minions.Get(name)
		if err != nil {
			return nil, err
		}

		minions = append(minions, minion)
	}

	return minions, nil
}

// GetQueueURLs returns an array of queue URLs. These URLs are extracted from
// the minions attached to this channel.
func (channelCfg *ChannelCfg) GetQueueURLs(srv *Server) ([]string, error) {
	var urls []string

	minions, err := channelCfg.GetMinions(srv)
	if err != nil {
		return urls, err
	}

	for _, minion := range minions {
		if minion.QueueURL == "" {
			log.Printf("minion without QueueURL: %s", minion.Name)
			continue
		}

		urls = append(urls, minion.QueueURL)
	}

	return urls, nil
}

// ParseConfigFile reads our JSON config file and validates its values, also
// populating defaults when possible.
func ParseConfigFile(cmd *CmdLine) (*Config, error) {
	cfg := &Config{}

	file, err := os.Open(cmd.ConfigFile)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	if cfg.QueueName == "" {
		return cfg, errors.New("'QueueName' is not defined")
	}

	if cfg.IRCServer == "" {
		return cfg, errors.New("'IRCServer' is not defined")
	}

	if cfg.IRCNickname == "" {
		return cfg, errors.New("'IRCNickname' is not defined")
	}

	if cfg.AWSRegionCode == "" {
		return cfg, errors.New("'AWSRegionCode' is not defined")
	}

	if cfg.AWSAccessKeyID == "" {
		return cfg, errors.New("'AWSAccessKeyID' is not defined")
	}

	if cfg.AWSSecretAccessKey == "" {
		return cfg, errors.New("'AWSSecretAccessKey' is not defined")
	}

	if cfg.AliasFilePath == "" {
		cfg.AliasFilePath = "aliases.cfg"
	}

	if cfg.MinionsFilePath == "" {
		cfg.MinionsFilePath = "minions.cfg"
	}

	return cfg, nil
}

// ParseCommandLine parses the command line arguments and populate the global
// cmd struct.
func ParseCommandLine() *CmdLine {
	cmd := &CmdLine{}
	flagParser := flags.NewParser(cmd, flags.PassDoubleDash)
	_, err := flagParser.Parse()
	if err != nil {
		println("command line error: " + err.Error())
		flagParser.WriteHelp(os.Stderr)
		os.Exit(1)
	}
	return cmd
}
