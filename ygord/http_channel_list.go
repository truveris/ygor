// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

// ChannelListHandler is an HTTP handler returning a JSON document with a list
// of all the registered channels to date.
type ChannelListHandler struct {
	*Server
}

type channel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ClientCount int    `json:"clientCount"`
}

type channelListResponse struct {
	Channels []channel `json:"channels"`
}

func (handler *ChannelListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := auth(r)
	if err != nil {
		errorHandler(w, "Authentication failed", err)
		return
	}

	response := channelListResponse{}

	// Strip the '#' from the channel, that identifier is given to the
	// tune-in handler.
	for name := range handler.Server.Config.Channels {
		ch := channel{
			ID:          strings.TrimPrefix(name, "#"),
			Name:        name,
			ClientCount: len(handler.Server.GetClientsByChannel(name)),
		}

		response.Channels = append(response.Channels, ch)
	}

	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)
	err = encoder.Encode(response)
	if err != nil {
		errorHandler(w, "failed to encode response JSON", err)
		return
	}
}
