// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
//
// Minions takes orders from ygord and executes them (through an SQS
// queue/inbox). There could be hundreds of minions installed on all sorts of
// machines, managed by ygord.
//
// Messages to minions should be nothing but plain text. They should take the
// form of a command and its parameters, for example:
//
// 	play valkyries.mp3
//
// The cost of one minion in SQS is less than a dollar a year, at one query
// per 20 seconds:
//
// 	Number of requests per day: (60 * 60 * 24) / 20 = 4320
// 	Number of requests per year: 4320 * 365 = 1576800
// 	Cost per request: $0.0000005
// 	Total cost per year: 1576800 * 0.0000005 = $0.7884
//

package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tamentis/go-mplayer"
	"github.com/truveris/sqs"
	"github.com/truveris/sqs/sqschan"
	"github.com/truveris/ygor"
)

var (
	// RunningProcess is a reference to the last process we launched. This
	// is useful to allow user to kill this process (e.g. stop the video).
	RunningProcess *os.Process
)

// startReceivingFromStdin is used for debugging. It fetches queue messages
// from stdin instead of AWS SQS.
func startReceivingFromStdin(incoming chan *sqs.Message) error {
	err := Register(cfg.Name, "fake-queue")
	if err != nil {
		return errors.New("registration failed: " + err.Error())
	}

	go func() {
		br := bufio.NewReader(os.Stdin)
		for {
			line, err := br.ReadString('\n')
			if err != nil {
				log.Fatal("terminating: " + err.Error())
			}
			line = strings.TrimSpace(line)

			incoming <- &sqs.Message{Body: line, UserID: "fakeUserID"}
		}
	}()

	return nil
}

// startReceivingFromSQS is the call used in a production system to start
// receiving messages. It is not used in test.
func startReceivingFromSQS(incoming chan *sqs.Message) error {
	client, err := sqs.NewClient(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey,
		cfg.AWSRegionCode)
	if err != nil {
		return err
	}

	attrs := sqs.CreateQueueAttributes{
		MaximumMessageSize:            4096,
		ReceiveMessageWaitTimeSeconds: 20,
		VisibilityTimeout:             10,
		MessageRetentionPeriod:        300,
	}

	queueURL, err := client.CreateQueueWithAttributes(cfg.QueueName, attrs)
	if err != nil {
		log.Fatal(err)
	}

	err = Register(cfg.Name, queueURL)
	if err != nil {
		return errors.New("registration failed: " + err.Error())
	}

	ch, errch, err := sqschan.IncomingFromURL(client, queueURL)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case err = <-errch:
				log.Printf("error reading sqs message: " + err.Error())
			case msg := <-ch:
				incoming <- msg
				client.DeleteMessage(msg)
			}
		}
	}()

	return nil
}

// SplitTwo separates command and data in various contexts.
func SplitTwo(body string) (string, string) {
	var command, data string

	tokens := strings.SplitN(body, " ", 2)

	command = tokens[0]
	if len(tokens) > 1 {
		data = tokens[1]
	}

	return command, data
}

// Send is used to send message to ygord. TODO: replace this by an sqschan.
func Send(message string) error {
	log.Printf("send to ygord: %s", message)
	if cfg.TestMode {
		return nil
	}

	if cfg.YgordQueueName == "" {
		return nil
	}

	client, err := sqs.NewClient(cfg.AWSAccessKeyID, cfg.AWSSecretAccessKey,
		cfg.AWSRegionCode)
	if err != nil {
		return err
	}

	url, err := client.GetQueueURL(cfg.YgordQueueName)
	if err != nil {
		return err
	}

	err = client.SendMessage(url, sqs.SQSEncode(message))
	if err != nil {
		return err
	}

	return nil
}

// Register sends a registration message to ygord: who we are and how to speak
// to us. TODO: this command should include capabilities (sound, turret, etc.)
func Register(name, queueURL string) error {
	message := fmt.Sprintf("register %s %s", cfg.Name, queueURL)
	err := Send(message)
	if err != nil {
		return err
	}

	return nil
}

// Loop until the end of time.
//
// In case of error, delay the next loop. Automatically reconnect if everything
// goes fine (for 0 or 1 message).
func main() {
	parseCommandLine()

	err := parseConfigFile()
	if err != nil {
		log.Fatal("config error: ", err.Error())
	}

	log.Printf("%s starting up", cfg.Name)

	// This is the message box.
	incoming := make(chan *sqs.Message)
	if cfg.TestMode {
		err := startReceivingFromStdin(incoming)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := startReceivingFromSQS(incoming)
		if err != nil {
			log.Fatal(err)
		}
	}

	go ygor.WaitForTraceRequest()

	go playNoise()

	if !cfg.TestMode {
		mplayer.StartSlave(mplayerErrorHandler)
		OpenTurrets()
	}

	for msg := range incoming {
		command, data := SplitTwo(sqs.SQSDecode(msg.Body))

		switch command {
		case "play", "play-tune":
			Play(data)
		case "say":
			Say(data)
		case "xombrero":
			Xombrero(data)
		case "shutup":
			ShutUp()
		case "reboot":
			Reboot()
		case "ping":
			Ping(data)
		case "volume":
			Volume(data)
		case "turret":
			Turret(data)
		case "error":
			// These errors are typically received when the queue
			// systems fails to fetch a message. There is no reason
			// at the moment for ygord to send errors to minions.
			log.Printf("error message: %s", data)
		case "register":
			log.Printf("registration: %s", data)
		default:
			log.Printf("unknown command: %s", msg)
		}
	}

	if !cfg.TestMode {
		CloseTurrets()
	}
}
