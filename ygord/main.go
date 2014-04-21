// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"bufio"
	"os"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/truveris/sqs"
	"github.com/truveris/sqs/sqschan"
	"github.com/truveris/ygor"
)

var (
	// All the incoming messages body and minions.
	inputQueue = make(chan *ygor.Message)

	// Everything coming from the minions.
	minionInbox = make(chan string)

	// All the modules currently registered.
	modules = make([]Module, 0)
)

type Module interface {
	Init()
}

func IRCMessageHandler(msg *ygor.Message) {
	for _, cmd := range ygor.RegisteredCommands {
		if !cmd.IRCMessageMatches(msg) {
			continue
		}

		if cmd.PrivMsgFunction == nil {
			Debug("unhandled IRC message: " + msg.Body)
			continue
		}

		cmd.PrivMsgFunction(msg)
		break
	}
}

func MinionMessageHandler(msg *ygor.Message) {
	for _, cmd := range ygor.RegisteredCommands {
		if !cmd.MinionMessageMatches(msg) {
			continue
		}

		if cmd.MinionMsgFunction == nil {
			Debug("unhandled minion message: " + msg.Body)
			continue
		}

		cmd.MinionMsgFunction(msg)
		break
	}
}

// Send a message to a channel.
func IRCPrivMsg(channel, msg string) {
	lines := strings.Split(msg, "\n")
	for i := 0; i < len(lines); i++ {
		if lines[i] == "" {
			continue
		}
		IRCOutgoing <- fmt.Sprintf("PRIVMSG %s :%s", channel, lines[i])

		// Make test mode faster.
		if cfg.TestMode {
			time.Sleep(50 * time.Millisecond)
		} else {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

// Send an action message to a channel.
func privAction(channel, msg string) {
	IRCOutgoing <- fmt.Sprintf("PRIVMSG %s :\x01ACTION %s\x01", channel, msg)
}

func delayedPrivMsg(channel, msg string, waitTime time.Duration) {
	time.Sleep(waitTime)
	IRCPrivMsg(channel, msg)
}

func delayedPrivAction(channel, msg string, waitTime time.Duration) {
	time.Sleep(waitTime)
	privAction(channel, msg)
}

func registerModule(module Module) {
	module.Init()
	modules = append(modules, module)
}

func joinChannel(channel string) {
	IRCOutgoing <- "JOIN " + channel
}

// Send the message to the configured debug channel if any.
func Debug(msg string) {
	if cfg.AdminChannel == "" {
		return
	}
	IRCPrivMsg(cfg.AdminChannel, msg)
}

// Auto-join all the configured channels.
func autojoin() {
	// Make test mode faster.
	if cfg.TestMode {
		time.Sleep(50 * time.Millisecond)
	} else {
		time.Sleep(500 * time.Millisecond)
	}

	for _, c := range cfg.GetAutoJoinChannels() {
		joinChannel(c)
	}
}

func NewMessageFromMinionLine(line string) *ygor.Message {
	msg := ygor.NewMessage()
	msg.Type = ygor.MsgTypeMinion
	msg.Body = line

	args := strings.Split(line, " ")
	msg.Command = args[0]
	msg.Args = args[1:]

	return msg
}

func NewMessageFromMinionSQS(sqsmsg *sqs.Message) *ygor.Message {
	msg := NewMessageFromMinionLine(sqsmsg.Body)
	msg.UserID = sqsmsg.UserID
	return msg
}

func startMinionHandlers(client *sqs.Client) (<-chan error, error) {
	errch := make(chan error, 0)

	ch, sqserrch, err := sqschan.Incoming(client, cfg.QueueName)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case sqsmsg := <-ch:
				msg := NewMessageFromMinionSQS(sqsmsg)
				inputQueue <- msg
				err := client.DeleteMessage(sqsmsg)
				if err != nil {
					errch <- err
				}
			case err := <-sqserrch:
				errch <- err
			}
		}
	}()

	return errch, nil
}

// This is used for debugging.
//
// It fetches queue messages from stdin instead of AWS SQS.
//
func StartStdinHandler() (chan error, chan error) {
	go func() {
		for {
			line := <-IRCOutgoing
			os.Stdout.WriteString(line + "\n")
		}
	}()

	go func() {
		br := bufio.NewReader(os.Stdin)
		for {
			line, err := br.ReadString('\n')
			if err != nil {
				log.Printf("terminating: " + err.Error())
				os.Exit(0)
			}
			line = strings.TrimSpace(line)

			args := strings.SplitN(line, " ", 2)
			if len(args) != 2 {
				log.Printf("not enough elements: %s", args)
				continue
			}

			msgtype := args[0]
			line = args[1]

			switch msgtype {
			case "irc":
				msg := NewMessageFromIRCLine(line)
				if msg != nil {
					inputQueue <- NewMessageFromIRCLine(line)
				}
			case "minion":
				args := strings.SplitN(line, " ", 2)
				if len(args) != 2 {
					log.Printf("not enough elements (minion): %s", args)
					continue
				}
				userid := args[0]
				line = args[1]
				msg := NewMessageFromMinionLine(line)
				msg.UserID = userid
				inputQueue <- msg
			default:
				log.Printf("unknown msg type: %s", msgtype)
			}
		}
	}()

	return nil, nil
}


func startHandlers() (<-chan error, <-chan error) {
	if cfg.TestMode {
		return StartStdinHandler()
	}

	client, err := sqs.NewClient(cfg.AWSAccessKeyId, cfg.AWSSecretAccessKey,
		cfg.AWSRegionCode)
	if err != nil {
		log.Fatal(err)
	}

	ircerrch, err := startIRCHandlers(client)
	if err != nil {
		log.Fatal("error starting IRC handler: "+err.Error())
	}

	minionerrch, err := startMinionHandlers(client)
	if err != nil {
		log.Fatal("error starting minion handler: " + err.Error())
	}

	return ircerrch, minionerrch
}

func main() {
	parseCommandLine()

	err := parseConfigFile()
	if err != nil {
		log.Fatal("config error: ", err.Error())
	}

	log.Printf("registering modules")

	registerModule(&AliasModule{})
	registerModule(&ImageModule{})
	registerModule(&RebootModule{})
	registerModule(&MinionsModule{})
	registerModule(&PingModule{})
	registerModule(&SayModule{})
	registerModule(&ShutUpModule{})
	registerModule(&SoundBoardModule{})
	registerModule(&XombreroModule{})

	log.Printf("starting i/o handlers")
	ircerrch, minionerrch := startHandlers()

	log.Printf("auto-join IRC channels")
	go autojoin()

	log.Printf("ready")
	for {
		select {
		case err := <-ircerrch:
			log.Printf("irc handler error: %s", err.Error())
		case err := <-minionerrch:
			log.Printf("minion handler error: %s", err.Error())
		case msg := <-inputQueue:
			switch msg.Type {
			case ygor.MsgTypeIRCChannel:
				IRCMessageHandler(msg)
			case ygor.MsgTypeIRCPrivate:
				IRCMessageHandler(msg)
			case ygor.MsgTypeMinion:
				MinionMessageHandler(msg)
			default:
				log.Printf("msg handler error: un-handled type '%d'", msg.Type)
			}
		}
	}

}
