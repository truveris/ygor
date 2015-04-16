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
	"github.com/thoj/go-ircevent"
	"log"
	"regexp"
	"strings"
)

var (
	// IRCOutgoing is the queue of all outgoing IRC messages.  All strings
	// going into this channel should formatted exactly as if they were
	// sent directly to the IRC server.
	IRCOutgoing   = make(chan string)
	IRCIncoming   = make(chan string)
	IRCDisconnect = make(chan string)
	conn          *irc.Connection

	reAddressed = regexp.MustCompile(`^(\w+)[:,.]*\s*(.*)`)
)

func expandSentence(words []string) ([][]string, error) {
	sentences := make([][]string, len(words))

	// Resolve any alias found as first word.
	expanded, err := Aliases.Resolve(words[0])
	if err != nil {
		return nil, err
	}

	sentences, err = LexerSplit(expanded)
	if err != nil {
		return nil, err
	}

	last := len(sentences) - 1
	sentences[last] = append(sentences[last], words[1:]...)

	return sentences, nil
}

// Expand sentences through aliases.
func expandSentences(ss [][]string) ([][]string, error) {
	sentences := make([][]string, len(ss))

	for _, words := range ss {
		if len(words) == 0 {
			continue
		}

		newsentences, err := expandSentence(words)
		if err != nil {
			return nil, err
		}
		sentences = append(sentences, newsentences...)
	}

	return sentences, nil
}

// NewMessagesFromBody creates a new ygor message from a plain string.
func NewMessagesFromBody(body string) ([]*Message, error) {
	var msgs []*Message

	sentences, err := LexerSplit(body)
	if err != nil {
		return nil, err
	}

	// TODO: make that recursive.
	for i := 0; i < 3; i++ {
		sentences, err = expandSentences(sentences)
		if err != nil {
			return nil, err
		}
	}

	for _, words := range sentences {
		if len(words) == 0 {
			continue
		}

		msg := NewMessage()

		msg.Body = strings.Join(words, " ")
		msg.Command = words[0]

		if len(words) > 1 {
			msg.Args = append(msg.Args, words[1:]...)
		}

		msgs = append(msgs, msg)
	}

	return msgs, nil
}

// NewMessagesFromPrivMsg create a new ygor message from a parsed PRIVMSG.
// func NewMessagesFromPrivMsg(privmsg *PrivMsg) []*Message {
// 	msgs, err := NewMessagesFromBody(privmsg.Body)
// 	if err != nil {
// 		if privmsg.Addressed {
// 			IRCPrivMsg(privmsg.ReplyTo, "lexer/expand error: "+
// 				err.Error())
// 		}
// 		return nil
// 	}
//
// 	for _, msg := range msgs {
// 		if privmsg.Direct {
// 			msg.Type = MsgTypeIRCPrivate
// 		} else {
// 			// Channel messages have to be prefixes with the bot's nick.
// 			if !privmsg.Addressed && !privmsg.Direct {
// 				return nil
// 			}
// 			msg.Type = MsgTypeIRCChannel
// 		}
//
// 		msg.UserID = privmsg.Nick
// 		msg.ReplyTo = privmsg.ReplyTo
// 	}
//
// 	return msgs
// }

// NewMessagesFromEvent creates a new array of messages based on a PRIVMSG event.
func NewMessagesFromEvent(e *irc.Event, nick string) []*Message {
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
	if tokens == nil || tokens[1] != nick {
		return nil
	}

	body := tokens[2]
	target := e.Arguments[0]

	// Sent directly to the bot, fuck that.  Everything is public.
	if target == nick {
		log.Printf("Ignoring private message: %s", e)
		return nil
	}

	msgs, err := NewMessagesFromBody(body)
	if err != nil {
		IRCPrivMsg(target, "lexer/expand error: "+err.Error())
		return nil
	}

	for _, msg := range msgs {
		msg.Type = MsgTypeIRCChannel
		msg.UserID = e.Nick
		msg.ReplyTo = target
	}

	return msgs
}

// NewMessagesFromIRCLine creates a new Message based on the raw IRC line.
// func NewMessagesFromIRCLine(line string) []*Message {
// 	privmsg := NewPrivMsg(line, cfg.IRCNickname)
// 	if privmsg == nil {
// 		// Not a PRIVMSG.
// 		return nil
// 	}
//
// 	// Check if we should ignore this message.
// 	for _, ignore := range cfg.Ignore {
// 		if ignore == privmsg.Nick {
// 			return nil
// 		}
// 	}
//
// 	return NewMessagesFromPrivMsg(privmsg)
// }

// IRCMessageHandler loops through the command registry to find a matching
// command and executes it.
func IRCMessageHandler(msg *Message) {
	for _, cmd := range RegisteredCommands {
		if !cmd.IRCMessageMatches(msg) {
			continue
		}

		if cmd.PrivMsgFunction == nil {
			log.Printf("misconfigured command: %s (no PrivMsg)",
				cmd.Name)
			continue
		}

		cmd.PrivMsgFunction(msg)
		return
	}

	// If we got that far, we didn't find a command.
	IRCPrivMsg(msg.ReplyTo, "command not found: "+msg.Command)
}

//
// All the functions below are used as wrappers to send data to the IRC server.
// They are convenience functions for ygor to speak.
//

// IRCPrivMsg sends a message to a channel. Construct a PRIVMSG and send the
// raw client line to the server.
func IRCPrivMsg(channel, msg string) {
	lines := strings.Split(msg, "\n")
	for i := 0; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}
		conn.Privmsg(channel, lines[i])
	}
}

// IRCPrivAction sends an action message to a channel.  This function is the
// equivalent of using the /ME command in a normal IRC client.
func IRCPrivAction(channel, msg string) {
	conn.Action(channel, msg)
}

// JoinChannel sends a /JOIN command to the server.
func JoinChannel(channel string) {
}

// Debug sends the message to the configured debug channel if any.
func Debug(msg string) {
	if cfg.AdminChannel == "" {
		return
	}
	IRCPrivMsg(cfg.AdminChannel, msg)
}

func StartIRCAdapter() error {
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
		msgs := NewMessagesFromEvent(e, cfg.IRCNickname)
		for _, msg := range msgs {
			IRCMessageHandler(msg)
		}
	})

	return nil
}
