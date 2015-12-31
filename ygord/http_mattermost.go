// Copyright 2015-2016, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// http_mattermost.go contains the API endpoint ingesting mattermost outgoing
// webhook requests.
//

package main

import (
	"encoding/json"
	"net/http"
)

// MattermostHandler is the HTTP Handler for the mattermost webhooks
type MattermostHandler struct {
	*Server
}

func (handler *MattermostHandler) ReplyToMattermost(w http.ResponseWriter, channel, text string) {
	response := handler.Server.NewMattermostResponse(channel, text)

	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)
	err := encoder.Encode(response)
	if err != nil {
		errorHandler(w, "failed to encode response JSON", err)
		return
	}
}

// ServeHTTP is a standard handler ServeHTTP request as expected by the
// standard http library.
func (handler *MattermostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := auth(r)
	if err != nil {
		errorHandler(w, "Authentication failed", err)
		return
	}

	if r.Method != "POST" {
		errorHandler(w, "Unsupported method", err)
		return
	}

	r.ParseForm()

	channelName := r.Form.Get("channel_name")
	if channelName == "" {
		errorHandler(w, "Mattermost channelName is required", err)
		return
	}

	srv := handler.Server

	token := r.Form.Get("token")
	if token != srv.Config.MattermostToken {
		errorHandler(w, "invalid Mattermost token", err)
		return
	}

	msgs := srv.NewMessagesFromMattermostRequest(r)
	for _, msg := range msgs {
		srv.InputQueue <- msg
	}
}
