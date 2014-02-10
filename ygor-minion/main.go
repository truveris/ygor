// Copyright 2014, Truveris Inc. All Rights Reserved.
//
// ygor-minion takes orders from ygor and executes them (through an SQS
// queue/inbox). There could be hundreds of minions installed on different
// machines, they can all have different purposes, it's up to ygor to decide.
//
// Messages to ygor-minion should be short and sweet, with nothing but plain
// text. They should take the form of a command and its parameters, for
// example:
//
// 	play-tune valkyries.mp3
//
// The cost of one ygor-minion in SQS is less than a dollar a year, at one
// query per 20 seconds:
//
// 	Number of requests per day: (60 * 60 * 24) / 20 = 4320
// 	Number of requests per year: 4320 * 365 = 1576800
// 	Cost per request: $0.0000005
// 	Total cost per year: 1576800 * 0.0000005 = $0.7884
//

package main

import (
	"encoding/xml"
	"errors"
	"github.com/mikedewar/aws4"
	"io/ioutil"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

const (
	APIVersion       = "2012-11-05"
	SignatureVersion = "4"
)

var QueueURL string


// struct defining what to extract from the XML document
type sqsMessage struct {
	Body          []string `xml:"ReceiveMessageResult>Message>Body"`
	ReceiptHandle []string `xml:"ReceiveMessageResult>Message>ReceiptHandle"`
}

func buildReceiveMessageURL() string {
	query := url.Values{}
	query.Set("Action", "ReceiveMessage")
	// query.Set("AttributeName", "All")
	query.Set("Version", APIVersion)
	query.Set("SignatureVersion", SignatureVersion)
	query.Set("WaitTimeSeconds", "20")
	query.Set("MaxNumberOfMessages", "1")
	url := QueueURL + "?" + query.Encode()
	return url
}

func buildDeleteMessageURL(receipt string) string {
	query := url.Values{}
	query.Set("Action", "DeleteMessage")
	query.Set("ReceiptHandle", receipt)
	query.Set("Version", APIVersion)
	query.Set("SignatureVersion", SignatureVersion)
	url := QueueURL + "?" + query.Encode()
	return url
}

// Return a client ready to use with the proper auth parameters.
func getClient() *aws4.Client {
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	if accessKey == "" {
		log.Fatal("missing AWS_ACCESS_KEY_ID")
	}

	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if secretKey == "" {
		log.Fatal("missing AWS_SECRET_ACCESS_KEY")
	}

	keys := &aws4.Keys{AccessKey: accessKey, SecretKey: secretKey}

	return &aws4.Client{Keys: keys}
}

// Return a single message body, with its ReceiptHandle. A lack of message is
// not considered an error but both strings will be empty.
func getMessage() (string, string, error) {
	var m sqsMessage

	client := getClient()
	url := buildReceiveMessageURL()

	resp, err := client.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != 200 {
		return "", "", errors.New(string(body))
	}

	err = xml.Unmarshal(body, &m)
	if err != nil {
		return "", "", err
	}

	// The API call is build to only return one or zero messages.
	if len(m.Body) < 1 {
		return "", "", nil
	}
	message := m.Body[0]
	receipt := m.ReceiptHandle[0]

	return message, receipt, nil
}

func deleteMessage(receipt string) error {
	client := getClient()
	url := buildDeleteMessageURL(receipt)

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}

func playTune(filename string) {
	filepath := "tunes/" + path.Base(filename)
	if _, err := os.Stat(filepath); err != nil {
		log.Printf("play-tune bad filename")
		return
	}

	err := exec.Command("mplayer", "-really-quiet", filepath).Run()
	if err != nil {
		log.Printf("error starting mplayer")
	}
}

// say (for macs)
func macSay(sentence string) {
	err := exec.Command("say", sentence).Run()
	if err != nil {
		log.Printf("error starting say")
	}
}

// espeak | aplay (for linux)
func say(sentence string) {
	var err error

	cmd_espeak := exec.Command("espeak", "-ven-us+f2", "--stdout",
		sentence, "-a", "300", "-s", "130")
	cmd_aplay := exec.Command("aplay")

	cmd_aplay.Stdin, err = cmd_espeak.StdoutPipe()
	if err != nil {
		log.Printf("error on cmd_espeak.StdoutPipe(): " + err.Error())
		return
	}

	err = cmd_espeak.Start()
	if err != nil {
		log.Printf("error on cmd_espeak.Start(): " + err.Error())
		return
	}
	err = cmd_aplay.Start()
	if err != nil {
		log.Printf("error on cmd_aplay.Start(): " + err.Error())
		return
	}

	err = cmd_espeak.Wait()
	if err != nil {
		log.Printf("error on cmd_espeak.Wait(): " + err.Error())
		return
	}
	err = cmd_aplay.Wait()
	if err != nil {
		log.Printf("error on cmd_aplay.Wait(): " + err.Error())
		return
	}
}


// Loop until the end of time.
//
// In case of error, delay the next loop. Automatically reconnect if everything
// goes fine (for 0 or 1 message).
func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: ygor-minion queue-url\n")
		os.Exit(1)
	}

	QueueURL = os.Args[1]

	log.Printf("ygor-minion ready!")

	for {
		body, receipt, err := getMessage()
		if err != nil {
			log.Printf("error: %s", err.Error())
			time.Sleep(10 * time.Second)
		}

		if body == "" {
			continue
		}

		log.Printf("got message: \"%s\"", body)
		deleteMessage(receipt)

		tokens := strings.Split(body, " ")
		switch tokens[0] {
		case "play-tune":
			playTune(tokens[1])
		case "mac-say":
			say(strings.Join(tokens[1:], " "))
		case "say":
			say(strings.Join(tokens[1:], " "))
		default:
			log.Printf("unknown command %s", tokens[0])
		}
	}
}
