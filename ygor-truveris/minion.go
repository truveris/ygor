// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

// TODO: replace the owner messages to DebugChannel.

package main

import (
	"fmt"
	"github.com/mikedewar/aws4"
	"io/ioutil"
	"net/url"
	"strings"
)

const (
	APIVersion       = "2012-11-05"
	SignatureVersion = "4"
	ContentType      = "application/x-www-form-urlencoded"
)

func buildSendMessageData(msg string) string {
	query := url.Values{}
	query.Set("Action", "SendMessage")
	query.Set("Version", APIVersion)
	query.Set("SignatureVersion", SignatureVersion)
	query.Set("MessageBody", msg)
	return query.Encode()
}

// Return a client ready to use with the proper auth parameters.
func getClient() *aws4.Client {
	keys := &aws4.Keys{
		AccessKey: cfg.AwsAccessKeyId,
		SecretKey: cfg.AwsSecretAccessKey,
	}
	return &aws4.Client{Keys: keys}
}

// Send a message to our friendly minion via its SQS queue.
func SendToMinion(channel, msg string) {
	client := getClient()
	data := buildSendMessageData(msg)
	channelCfg, exists := cfg.Channels[channel]
	if !exists {
		Debug("error: "+channel+" has no queue(s) configured")
		return
	}

	if cfg.Debug {
		fmt.Printf("[SQS-SendToMinion] %s\n", msg)
		return
	}

	resp, err := client.Post(channelCfg.QueueURL, ContentType,
		strings.NewReader(data))
	if err != nil {
		Debug("error sending to minion: "+err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Debug("error sending to minion: "+err.Error())
		return
	}

	if resp.StatusCode != 200 {
		Debug("error sending to minion: "+string(body))
		return
	}
}

