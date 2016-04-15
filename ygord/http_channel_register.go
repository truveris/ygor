// Copyright 2014-2016, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// ChannelRegisterHandler is the HTTP handler for clients to register to a
// channel.
type ChannelRegisterHandler struct {
	*Server
}

type channelRegisterRequest struct {
	ChannelID string `json:"channelID"`
}

type channelRegisterResponse struct {
	ClientID string `json:"clientID"`
}

func (handler *ChannelRegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, err := auth(r)
	if err != nil {
		errorHandler(w, "Authentication failed", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	input := &channelRegisterRequest{}
	err = decoder.Decode(input)
	if err != nil {
		errorHandler(w, "Failed to decode input JSON", err)
		return
	}

	// Extract the IP address, attempting first to get it from the
	// X-Forwarded-For header which is generally populated by reverse
	// proxies.
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = strings.SplitN(r.RemoteAddr, ":", 2)[0]
	}

	client := handler.Server.RegisterClient(username, input.ChannelID)
	client.IPAddress = ip
	if agent, ok := r.Header["User-Agent"]; ok {
		client.UserAgent = agent[0]
	}

	w.Header().Set("Content-Type", "application/json")

	select {
	case <-time.After(time.Second * 2):
	}

	jsonHandler(w, channelRegisterResponse{ClientID: client.ID})
}
