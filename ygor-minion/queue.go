// Copyright 2014, Truveris Inc. All Rights Reserved.

package main

import (
	"encoding/xml"
	"errors"
	"github.com/mikedewar/aws4"
	"io/ioutil"
	"net/url"
)

const (
	APIVersion       = "2012-11-05"
	SignatureVersion = "4"
)

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
	url := cfg.QueueURL + "?" + query.Encode()
	return url
}

func buildDeleteMessageURL(receipt string) string {
	query := url.Values{}
	query.Set("Action", "DeleteMessage")
	query.Set("ReceiptHandle", receipt)
	query.Set("Version", APIVersion)
	query.Set("SignatureVersion", SignatureVersion)
	url := cfg.QueueURL + "?" + query.Encode()
	return url
}

// Return a client ready to use with the proper auth parameters.
func getClient() *aws4.Client {
	keys := &aws4.Keys{
		AccessKey: cfg.AwsAccessKeyId,
		SecretKey: cfg.AwsSecretAccessKey,
	}
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
