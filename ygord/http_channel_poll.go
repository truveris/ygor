// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// ChannelPollHandler is the HTTP handler user by clients to wait on content
// from the ygor server.  Client post a JSON object containing their ClientID
// and receive the array of commands to execute.
type ChannelPollHandler struct {
	*Server
}

type channelPollRequest struct {
	ClientID string
}

type channelPollResponse struct {
	Status   string
	Commands []string
}

func (handler *ChannelPollHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := auth(r)
	if err != nil {
		errorHandler(w, "Authentication failed", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	input := &channelPollRequest{}
	err = decoder.Decode(input)
	if err != nil {
		errorHandler(w, "Failed to decode input JSON", err)
		return
	}

	client := handler.Server.GetClientFromID(input.ClientID)
	if client == nil {
		jsonHandler(w, channelPollResponse{Status: "unknown-client"})
		return
	}

	client.KeepAlive()

	response := channelPollResponse{}

	// First try to get all the commands in the queue.
pullChan:
	for {
		select {
		case msg, ok := <-client.Queue:
			if ok {
				response.Status = "command"
				response.Commands = append(response.Commands, msg)
			} else {
				response.Status = "closed"
				goto end
			}
		default:
			break pullChan
		}
	}

	// If we didn't find any, just wait a few seconds.
	if len(response.Commands) == 0 {
		select {
		case msg := <-client.Queue:
			response.Status = "command"
			response.Commands = append(response.Commands, msg)
		case <-time.After(time.Second * 20):
			response.Status = "empty"
		}
	}

end:
	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)
	err = encoder.Encode(response)
	if err != nil {
		errorHandler(w, "failed to encode response JSON", err)
		return
	}
}
