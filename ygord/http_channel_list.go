// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

type ChannelListHandler struct {
	*Server
}

type Channel struct {
	Id   string
	Name string
}

type ChannelListResponse struct {
	Channels []Channel
}

func (handler *ChannelListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := auth(r)
	if err != nil {
		errorHandler(w, "Authentication failed", err)
		return
	}

	response := ChannelListResponse{}

	// Strip the '#' from the channel, that identifier is given to the
	// tune-in handler.
	for name, _ := range handler.Server.Config.Channels {
		response.Channels = append(response.Channels, Channel{
			Id:   strings.TrimPrefix(name, "#"),
			Name: name,
		})
	}

	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)
	err = encoder.Encode(response)
	if err != nil {
		errorHandler(w, "failed to encode response JSON", err)
		return
	}
}
