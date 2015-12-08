// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// The io_irc_http is a temporary hack to expose ygor functions to the
// interwebs. This is bound to be replaced by an external process communicating
// with ygord via the SQS "API".
//

package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
)

// Given a Basic Authorization header value, return the user and password.
func parseBasicAuth(value string) (string, string, error) {
	authorization, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", "", err
	}

	tokens := strings.SplitN(string(authorization), ":", 2)
	if len(tokens) != 2 {
		return "", "", errors.New("Unable to split Basic Auth.")
	}

	return tokens[0], tokens[1], nil
}

// Makes sure we have a Basic auth user. We don't check the password, we assume
// this HTTP server sits behind a proxy which enforces that aspect.
func auth(r *http.Request) (string, error) {
	var user string
	var err error

	auth, ok := r.Header["Authorization"]
	if ok {
		if len(auth) > 0 {
			if !strings.HasPrefix(auth[0], "Basic ") {
				return "", errors.New("Unsupported auth type")
			}
			value := strings.TrimPrefix(auth[0], "Basic ")
			user, _, err = parseBasicAuth(value)
			if err != nil {
				return "", err
			}
		}
	}

	return user, nil
}

func errorHandler(w http.ResponseWriter, msg string, err error) {
	if err != nil {
		log.Printf("%s: %s", msg, err.Error())
	} else {
		log.Printf("%s", msg)
	}

	w.Header().Set("Content-Type", "text/html")

	http.Error(w, msg, 500)
}

func jsonHandler(w http.ResponseWriter, obj interface{}) {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(obj)
	if err != nil {
		errorHandler(w, "failed to encode response JSON", err)
		return
	}
}

// HTTPServer starts an HTTP server at the given address.  This is started as a
// go routine by StartHTTPServer.
func HTTPServer(srv *Server, address string) {
	log.Printf("starting http server on %s", address)

	http.Handle("/", http.FileServer(http.Dir(srv.Config.WebRoot)))
	http.Handle("/alias/list", &AliasListHandler{srv})
	http.Handle("/channel/list", &ChannelListHandler{srv})
	http.Handle("/channel/register", &ChannelRegisterHandler{srv})
	http.Handle("/channel/poll", &ChannelPollHandler{srv})
	http.Handle("/client/list", &ClientListHandler{srv})

	err := http.ListenAndServe(address, nil)
	if err != nil {
		log.Fatal("unable to listen for http server", err)
	}
}

// StartHTTPServer starts an HTTP server routine if an address is configured.
func (srv *Server) StartHTTPServer(address string) error {
	if address != "" {
		go HTTPServer(srv, address)
	}
	return nil
}
