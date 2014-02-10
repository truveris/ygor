// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
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
func sendToMinion(msg string) {
	client := getClient()
	data := buildSendMessageData(msg)

	resp, err := client.Post(cfg.QueueURL, ContentType,
		strings.NewReader(data))
	if err != nil {
		privMsg(owner, err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		privMsg(owner, err.Error())
	}

	if resp.StatusCode != 200 {
		privMsg(owner, string(body))
	}
}
