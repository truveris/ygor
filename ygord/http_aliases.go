// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// http_aliases.go contains the API endpoint returning the list of aliases.
//

package main

import (
	"encoding/json"
	"github.com/truveris/ygor/ygord/alias"
	"net/http"
)

// AliasListHandler is the HTTP Handler for the list of aliases.
type AliasListHandler struct {
	*Server
}

// AliasListResponse is the struct returned as JSON in response to a request
// on this endpoint.
type AliasListResponse struct {
	Aliases []alias.Alias `json:"aliases"`
}

// ServeHTTP is a standard handler ServeHTTP request as expected by the
// standard http library.
func (handler *AliasListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := auth(r)
	if err != nil {
		errorHandler(w, "Authentication failed", err)
		return
	}

	aliases, err := handler.Server.Aliases.All()
	if err != nil {
		errorHandler(w, "failed to get aliases", err)
		return
	}

	response := AliasListResponse{Aliases: aliases}

	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)
	err = encoder.Encode(response)
	if err != nil {
		errorHandler(w, "failed to encode response JSON", err)
		return
	}
}
