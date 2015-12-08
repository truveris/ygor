// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
	"time"
)

// ClientListHandler is an HTTP handler returning a JSON document with a list
// of all the registered channels to date.
type ClientListHandler struct {
	*Server
}

type respClient struct {
	Username  string    `json:"username"`
	Channel   string    `json:"channel"`
	UserAgent string    `json:"userAgent"`
	IPAddress string    `json:"ipAddress"`
	LastSeen  time.Time `json:"lastSeen"`
}

type respClientList struct {
	Clients []respClient `json:"clients"`
}

func (handler *ClientListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := auth(r)
	if err != nil {
		errorHandler(w, "Authentication failed", err)
		return
	}

	response := respClientList{}

	for _, client := range handler.Server.ClientRegistry {
		response.Clients = append(response.Clients, respClient{
			Username:  client.Username,
			Channel:   client.Channel,
			UserAgent: client.UserAgent,
			IPAddress: client.IPAddress,
			LastSeen:  client.LastSeen,
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
