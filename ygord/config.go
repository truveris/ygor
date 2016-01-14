// Copyright 2014-2016, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/jessevdk/go-flags"
)

// CmdLine is a singleton used to store the command-line parameters.
type CmdLine struct {
	ConfigFile string `short:"c" description:"Configuration file" default:"/etc/ygord.conf"`
}

// ChannelCfg represents a per-channel grouping of minions.
type ChannelCfg struct {
}

// Config is a singleton used to store the file configuration.
type Config struct {
	// All the configured channels. ygord will JOIN every single one of
	// them and will push commands to the configured associated minions.
	Channels map[string]ChannelCfg

	// Hostname and port to use to connect to the IRC server.
	IRCServer string

	// Nickname of the bot. FIXME: this is not currently synchronized
	Nickname string

	// Try to send debug information to this channel in lieu of log file.
	AdminChannel string

	// Any chatter from these nicks will be dropped (other bots).
	Ignore []string

	// Where to find the alias file. Will use the local alias file found in
	// the current directory by default.
	AliasFilePath string

	// Where to find the web files (static folder).
	WebRoot string

	// If defined, start a web server to list the aliases (e.g. :8989)
	HTTPServerAddress string

	// If defined, it enables the "say" command and converts sentences into
	// streamable sound bites via a minion-accessible sayd.
	SaydURL string

	// If defined, it allows SoundCloud URLs to be resolved when passing URLs
	// to commands that utilize MediaObj.
	SoundCloudClientID string

	// If that value exist and is non-zero, ygor will automatically run the
	// "screensaver" alias if it exists.
	ScreensaverDelay int

	// Mattermost configuration.
	MattermostToken    string
	MattermostIconURL  string
	MattermostUsername string
	MattermostWebhook  string
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

	if cfg.Nickname == "" {
		return cfg, errors.New("'Nickname' is not defined")
	}

	if cfg.AliasFilePath == "" {
		cfg.AliasFilePath = "aliases.cfg"
	}

	// If a web server is started, make sure we configure a web root.
	if cfg.HTTPServerAddress != "" {
		if cfg.WebRoot == "" {
			return cfg, errors.New("'WebRoot' is not defined")
		}
	}

	// No delay configured == 15 minutes
	if cfg.ScreensaverDelay == 0 {
		cfg.ScreensaverDelay = 900
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
