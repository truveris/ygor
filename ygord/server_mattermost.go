// Copyright 2015-2016, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This server_mattermost file contains the mattermost server code, responsible
// for converting mattermost http Request into ygor messages.
//
// The message in this adapter is roughly converted as such:
//
//     Mattermost Server -> ygor webhook -> go http.Request -> ygor.Message
//

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type MattermostResponse struct {
	Text     string `json:"text"`
	Channel  string `json:"channel"`
	IconURL  string `json:"icon_url"`
	Username string `json:"username"`
}

func (srv *Server) NewMattermostResponse(channel, text string) *MattermostResponse {
	return &MattermostResponse{
		Text:     text,
		Channel:  channel,
		IconURL:  srv.Config.MattermostIconURL,
		Username: srv.Config.MattermostUsername,
	}
}

func (srv *Server) SendToMattermost(response *MattermostResponse) {
	client := &http.Client{}
	buf, err := json.Marshal(response)
	if err != nil {
		log.Printf("SendToMattermost: invalid JSON: %s", err.Error())
		return
	}
	resp, err := client.Post(srv.Config.MattermostWebhook,
		"application/json", bytes.NewReader(buf))
	if err != nil {
		log.Printf("SendToMattermost: http error: %s", err.Error())
		return
	}
	defer resp.Body.Close()

	// Read entire body to completion to re-use keep-alive connections.
	io.Copy(ioutil.Discard, resp.Body)

	if resp.StatusCode != 200 {
		log.Printf("SendToMattermost: unexpected status code: %d",
			resp.StatusCode)
	}
}

// NewMessagesFromMattermostRequest creates a new array of messages based on a
// PRIVMSG event.
func (srv *Server) NewMessagesFromMattermostRequest(r *http.Request) []*InputMessage {
	cfg := srv.Config

	// Check if we should ignore this message.
	for _, ignore := range cfg.Ignore {
		if ignore == r.Form.Get("userName") {
			log.Printf("Ignoring %s", ignore)
			return nil
		}
	}

	// Ignore the message if not prefixed with our nickname.  If it is,
	// remove this prefix from the body of the message.
	tokens := reAddressed.FindStringSubmatch(r.Form.Get("text"))
	if tokens == nil || tokens[1] != cfg.Nickname {
		return nil
	}

	body := strings.TrimSpace(tokens[2])
	target := r.Form.Get("channel_name")

	msgs, err := srv.NewMessagesFromBody(body, 0)
	if err != nil {
		srv.SendToMattermost(srv.NewMattermostResponse(target,
			"lexer/expand error: "+err.Error()))
		return nil
	}

	for _, msg := range msgs {
		msg.Type = InputMsgTypeMattermost
		msg.Nickname = r.Form.Get("user_name")
		msg.ReplyTo = target
	}

	return msgs
}
