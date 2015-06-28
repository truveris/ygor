// Copyright 2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
)

type connectHandler struct {
	*Server
}

type connectResponse struct {
	Status string
	Queue  string
}

func (handler *connectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := auth(r)
	if err != nil {
		errorHandler(w, "Authentication failed", err)
		return
	}

	if r.Method != "POST" {
		errorHandler(w, "connect is POST only", nil)
		return
	}

	encoder := json.NewEncoder(w)
	encoder.Encode(connectResponse{
		Status: "success",
		Queue:  "qweqweqwe",
	})
}
