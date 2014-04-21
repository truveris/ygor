// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"

	"github.com/truveris/sqs"
	"github.com/truveris/ygor"
)

func SendToQueueUsingClient(client *sqs.Client, queueURL, msg string) {
	if cfg.TestMode {
		fmt.Printf("[SQS-SendToMinion] %s %s\n", queueURL, msg)
		return
	}

	err := client.SendMessage(queueURL, msg)
	if err != nil {
		Debug("error sending to minion: " + err.Error())
	}
}

// Return all the minions configured for that channel.
func GetChannelMinions(channel string) []*ygor.Minion {
	channelCfg, exists := cfg.Channels[channel]
	if !exists {
		Debug("error: " + channel + " has no queue(s) configured")
		return nil
	}

	minions, err := channelCfg.GetMinions()
	if err != nil {
		Debug("error: GetChannelMinions: " + err.Error())
	}

	return minions
}

// Send a message to our friendly minion via its SQS queue.
func SendToChannelMinions(channel, msg string) {
	client, err := sqs.NewClient(cfg.AWSAccessKeyId, cfg.AWSSecretAccessKey,
		cfg.AWSRegionCode)
	if err != nil {
		Debug("error: " + err.Error())
		return
	}

	channelCfg, exists := cfg.Channels[channel]
	if !exists {
		Debug("error: " + channel + " has no queue(s) configured")
		return
	}

	urls, err := channelCfg.GetQueueURLs()
	if err != nil {
		Debug("error: unable to load queue URLs, " + err.Error())
		return
	}

	// Send the same exact data to all this channel's minion.
	for _, url := range urls {
		if cfg.TestMode {
			fmt.Printf("[SQS-SendToMinion] %s %s\n", url, msg)
			continue
		}

		err := client.SendMessage(url, msg)
		if err != nil {
			Debug("error sending to minion: " + err.Error())
			continue
		}
	}
}

// Send a message to our friendly minion via its SQS queue.
func SendToQueue(queueURL, msg string) error {
	client, err := sqs.NewClient(cfg.AWSAccessKeyId, cfg.AWSSecretAccessKey,
		cfg.AWSRegionCode)
	if err != nil {
		return err
	}
	SendToQueueUsingClient(client, queueURL, msg)

	return nil
}
