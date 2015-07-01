// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// A Client is a single web browser connected and registered to the server.
// The registry of clients is kept in memory and lost every time the server
// restarts.  Clients are responsible for reconnecting.
//

package main

import (
	"crypto/sha512"
	"fmt"
	"time"
)

const (
	MaxQueueLength = 1024
)

// Client represents a single ygor client in memory.
type Client struct {
	Username string
	Channel  string
	ID       string
	Queue    chan string
	LastSeen time.Time
}

// IsAlive checks if the client is still accepting messages. It will return
// false if the queue is full or if the client has been silent for too long.
func (c *Client) IsAlive() bool {
	if len(c.Queue) >= MaxQueueLength {
		return false
	}
	old := time.Now().Add(time.Hour * -48)
	if c.LastSeen.Before(old) {
		return false
	}
	return true
}

func (c *Client) KeepAlive() {
	c.LastSeen = time.Now()
}

// RegisterClient generates a new ID for this client, using the server salt and
// the current time baked into a SHA512 in an attempt to make this identified
// hard to predict.
func (srv *Server) RegisterClient(username, channel string) string {
	hash := sha512.New()
	hash.Write([]byte(fmt.Sprintf("%s%s%d", username, channel, time.Now().UnixNano())))
	hash.Write(srv.Salt)

	ID := fmt.Sprintf("%x", hash.Sum(nil))

	srv.ClientRegistry[ID] = &Client{
		Username: username,
		Channel:  channel,
		ID:       ID,
		Queue:    make(chan string, MaxQueueLength),
		LastSeen: time.Now(),
	}
	return ID
}

func (srv *Server) GetClientFromID(ID string) *Client {
	client, ok := srv.ClientRegistry[ID]
	if !ok {
		return nil
	}

	return client
}

func (srv *Server) GetClientsByChannel(channel string) []*Client {
	var clients []*Client

	for _, client := range srv.ClientRegistry {
		if client.Channel == channel {
			clients = append(clients, client)
		}
	}

	return clients
}

func (srv *Server) PurgeClient(client *Client) {
}
