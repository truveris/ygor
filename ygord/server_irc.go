// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// The io_irc portion of the ygor code base defines the adapter used to receive
// and send data to IRC.
//
// The message in this adapter is roughly converted as such:
//
//     IRC -> ygor.PrivMsg -> ygor.Message
//

package main

import (
	"log"
	"regexp"
	"strings"

	"github.com/thoj/go-ircevent"
	"github.com/truveris/ygor/ygord/lexer"
)

var (
	conn        *irc.Connection
	reAddressed = regexp.MustCompile(`^\s*(\w+)[:,.]*\s*(.*)`)
)

// NewMessagesFromBody creates a new ygor message from a plain string.
func (srv *Server) NewMessagesFromBody(body string, recursion int) ([]*IRCInputMessage, error) {
	var msgs []*IRCInputMessage

	sentences, err := lexer.Split(body)
	if err != nil {
		return nil, err
	}

	for i := 0; i < 3; i++ {
		sentences, err = srv.Aliases.ExpandSentences(sentences, recursion)
		if err != nil {
			return nil, err
		}
	}

	for _, words := range sentences {
		if len(words) == 0 {
			continue
		}

		msg := NewIRCInputMessage()
		msg.Recursion = recursion + 1
		msg.Body = strings.Join(words, " ")
		msg.Command = words[0]

		if len(words) > 1 {
			msg.Args = append(msg.Args, words[1:]...)
		}

		msgs = append(msgs, msg)
	}

	return msgs, nil
}

// NewMessagesFromIRCEvent creates a new array of messages based on a PRIVMSG event.
func (srv *Server) NewMessagesFromIRCEvent(e *irc.Event) []*IRCInputMessage {
	cfg := srv.Config

	// Check if we should ignore this message.
	for _, ignore := range cfg.Ignore {
		if ignore == e.Nick {
			log.Printf("Ignoring %s", ignore)
			return nil
		}
	}

	// Ignore the message if not prefixed with our nickname.  If it is,
	// remove this prefix from the body of the message.
	tokens := reAddressed.FindStringSubmatch(e.Message())
	if tokens == nil || tokens[1] != cfg.IRCNickname {
		return nil
	}

	body := strings.TrimSpace(tokens[2])
	target := e.Arguments[0]

	// Sent directly to the bot, fuck that.  Everything is public.
	if target == cfg.IRCNickname {
		log.Printf("Ignoring private message: %s", e)
		return nil
	}

	msgs, err := srv.NewMessagesFromBody(body, 0)
	if err != nil {
		srv.IRCPrivMsg(target, "lexer/expand error: "+err.Error())
		return nil
	}

	for _, msg := range msgs {
		msg.Type = IRCInputMsgTypeIRCChannel
		msg.Nickname = e.Nick
		msg.ReplyTo = target
	}

	return msgs
}

// IRCMessageHandler loops through the command registry to find a matching
// command and executes it.
func (srv *Server) IRCMessageHandler(msg *IRCInputMessage) {
	for _, cmd := range srv.RegisteredCommands {
		if !cmd.IRCMessageMatches(srv, msg) {
			continue
		}

		if cmd.PrivMsgFunction == nil {
			log.Printf("misconfigured command: %s (no PrivMsg)",
				cmd.Name)
			continue
		}

		log.Printf("cmd.PrivMsgFunction %s (rec:%d)", cmd.Name,
			msg.Recursion)
		cmd.PrivMsgFunction(srv, msg)
		return
	}

	// If we got that far, we didn't find a command.
	srv.IRCPrivMsg(msg.ReplyTo, "command not found: "+msg.Command)
}

//
// All the functions below are used as wrappers to send data to the IRC server.
// They are convenience functions for ygor to speak.
//

// IRCPrivMsg sends a message to a channel. Construct a PRIVMSG and send the
// raw client line to the server.
func (srv *Server) IRCPrivMsg(channel, msg string) {
	lines := strings.Split(msg, "\n")
	for i := 0; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}
		outmsg := &IRCOutputMessage{
			Type:    IRCOutMsgTypePrivMsg,
			Channel: channel,
			Body:    lines[i],
		}
		srv.IRCOutputQueue <- outmsg
	}
}

// IRCPrivAction sends an action message to a channel.  This function is the
// equivalent of using the /ME command in a normal IRC client.
func (srv *Server) IRCPrivAction(channel, msg string) {
	outmsg := &IRCOutputMessage{
		Type:    IRCOutMsgTypeAction,
		Channel: channel,
		Body:    msg,
	}
	srv.IRCOutputQueue <- outmsg
}

// StartIRCClient connects the server to the IRC server.
func (srv *Server) StartIRCClient() error {
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
		msgs := srv.NewMessagesFromIRCEvent(e)
		for _, msg := range msgs {
			srv.IRCInputQueue <- msg
		}
	})

	return nil
}
