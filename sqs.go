// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package ygor

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/url"
	"net/http"
	"strings"

	"github.com/mikedewar/aws4"
)

const (
	SQSAPIVersion       = "2012-11-05"
	SQSSignatureVersion = "4"
	SQSContentType      = "application/x-www-form-urlencoded"
)

type SQSClient struct {
	Aws4Client *aws4.Client
}

// struct defining what to extract from the XML document received in response
// to the GetMessage API call.
type SQSMessage struct {
	Body          []string `xml:"ReceiveMessageResult>Message>Body"`
	ReceiptHandle []string `xml:"ReceiveMessageResult>Message>ReceiptHandle"`
}

// Build the data portion of a Message POST API call.
func (client *SQSClient) BuildSendMessageData(msg string) string {
	query := url.Values{}
	query.Set("Action", "SendMessage")
	query.Set("Version", SQSAPIVersion)
	query.Set("SignatureVersion", SQSSignatureVersion)
	query.Set("MessageBody", msg)
	return query.Encode()
}

// Build the URL to conduct a ReceiveMessage GET API call.
func (client *SQSClient) BuildReceiveMessageURL(queueURL string) string {
	query := url.Values{}
	query.Set("Action", "ReceiveMessage")
	// query.Set("AttributeName", "All")
	query.Set("Version", SQSAPIVersion)
	query.Set("SignatureVersion", SQSSignatureVersion)
	query.Set("WaitTimeSeconds", "20")
	query.Set("MaxNumberOfMessages", "1")
	url := queueURL + "?" + query.Encode()
	return url
}

// Build the URL to conduct a DeleteMessage GET API call.
func (client *SQSClient) BuildDeleteMessageURL(queueURL, receipt string) string {
	query := url.Values{}
	query.Set("Action", "DeleteMessage")
	query.Set("ReceiptHandle", receipt)
	query.Set("Version", SQSAPIVersion)
	query.Set("SignatureVersion", SQSSignatureVersion)
	url := queueURL + "?" + query.Encode()
	return url
}

// Simple wrapper around the aws4 client Post() but less verbose.
func (client *SQSClient) Post(queueURL, data string) (*http.Response, error) {
	return client.Aws4Client.Post(queueURL, SQSContentType,
		strings.NewReader(data))
}

// Simple wrapper around the aws4 Get() to keep it consistent.
func (client *SQSClient) Get(url string) (*http.Response, error) {
	return client.Aws4Client.Get(url)
}

// Return a client ready to use with the proper auth parameters.
func NewSQSClient(awsAccessKeyId, awsSecretAccessKey string) *SQSClient {
	keys := &aws4.Keys{
		AccessKey: awsAccessKeyId,
		SecretKey: awsSecretAccessKey,
	}
	return &SQSClient{Aws4Client: &aws4.Client{Keys: keys}}
}

// Return a single message body, with its ReceiptHandle. A lack of message is
// not considered an error but both strings will be empty.
func (client *SQSClient) GetMessage(queueURL string) (string, string, error) {
	var m SQSMessage

	url := client.BuildReceiveMessageURL(queueURL)

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

// Conduct a DeleteMessage API call on the given queue, using the receipt
// handle from a previously fetched message.
func (client *SQSClient) DeleteMessage(queueURL, receipt string) error {
	url := client.BuildDeleteMessageURL(queueURL, receipt)

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	resp.Body.Close()

	return nil
}
