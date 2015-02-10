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
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func aliasesTxtHandler(w http.ResponseWriter, r *http.Request) {
	_, err := auth(r)
	if err != nil {
		log.Printf("Authentication failed: %s", err.Error())
		errorHandler(w, "Authentication failed")
		return
	}

	aliases, err := Aliases.All()
	if err != nil {
		http.Error(w, "error: "+err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	for _, alias := range aliases {
		fmt.Fprintf(w, "%s\t%s\n", alias.Name, alias.Value)
	}
}

func aliasesHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth(r)
	if err != nil {
		log.Printf("Authentication failed: %s", err.Error())
		errorHandler(w, "Authentication failed")
		return
	}

	aliases, err := Aliases.All()
	if err != nil {
		http.Error(w, "error: "+err.Error(), 500)
		return
	}

	fmt.Fprintf(w, `
	<html>
		<head>
			<title>ygor - aliases</title>
			<style type="text/css">
				body { font-family: monospace; }
				th { text-align: left; }
				th, td { padding: 2px 8px; }
			</style>
		</head>
		<body>
			<h1>%s@ygor/aliases</h1>
			<table>
				<thead>
					<tr>
						<th>Name</th>
						<th>Value</th>
					</tr>
				</thead>
				<tbody>
	`, user)

	re := regexp.MustCompile("(https?://(?:(?:[:&=+$,a-zA-Z0-9_-]+@)?[a-zA-Z0-9.-]+(?::[0-9])?)(?:/[,:!+=~%/.a-zA-Z0-9_()-]*)?\\??(?:[,:!+=&%@.a-zA-Z0-9_()-]*))")

	for _, alias := range aliases {

		value := re.ReplaceAll([]byte(alias.Value), []byte("<a href='$1'>$1</a>"))

		fmt.Fprintf(w, `
		<tr>
			<td>%s</td>
			<td>%s</td>
		</tr>`, alias.Name, value)
	}

	fmt.Fprintf(w, `
		</body>
	</html>
	`)
}

func minionsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth(r)
	if err != nil {
		log.Printf("Authentication failed: %s", err.Error())
		errorHandler(w, "Authentication failed")
		return
	}

	minions, err := Minions.All()
	if err != nil {
		http.Error(w, "error: "+err.Error(), 500)
		return
	}

	fmt.Fprintf(w, `
	<html>
		<head>
			<title>ygor - minions</title>
			<style type="text/css">
				body { font-family: monospace; }
				th { text-align: left; }
				th, td { padding: 2px 8px; }
			</style>
		</head>
		<body>
			<h1>%s@ygor/minions</h1>
			<table>
				<thead>
					<tr>
						<th>Name</th>
						<th>Last registration</th>
					</tr>
				</thead>
				<tbody>
	`, user)

	for _, minion := range minions {
		fmt.Fprintf(w, `
		<tr>
			<td>%s</td>
			<td>%s</td>
		</tr>`, minion.Name, minion.LastSeen)
	}

	fmt.Fprintf(w, `
		</body>
	</html>
	`)
}

// Given a Basic Authorization header value, return the user and password.
func parseBasicAuth(value string) (string, string, error) {
	log.Printf("parseBasicAuth: %s", value)
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

func errorHandler(w http.ResponseWriter, msg string) {
	fmt.Fprintf(w, `
	<html>
	<head><title>ygor: Error</title></head>
	<body>
		<h1>Error: %s</h1>
	</body>
	</html>
	`, msg)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth(r)
	if err != nil {
		log.Printf("Authentication failed: %s", err.Error())
		errorHandler(w, "Authentication failed")
		return
	}

	fmt.Fprintf(w, `
	<html>
	<head>
		<title>ygor</title>
		<style type="text/css">
			body { font-family: monospace; }
			th { text-align: left; }
			th, td { padding: 2px 8px; }
		</style>
	</head>
	<body>
		<h1>%s@ygor</h1>
		<ul>
			<li><a href="/aliases">aliases</a>
			<li><a href="/minions">minions</a>
		</ul>
	</body>
	</html>
	`, user)
}

// HTTPServer starts an HTTP server at the given address.  This is started as a
// go routine by StartHTTPAdapter.
func HTTPServer(address string) {
	log.Printf("starting http server on %s", address)
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/aliases.txt", aliasesTxtHandler)
	http.HandleFunc("/aliases", aliasesHandler)
	http.HandleFunc("/minions", minionsHandler)
	http.ListenAndServe(address, nil)
}

// StartHTTPAdapter starts an HTTP server routine if an address is configured.
func StartHTTPAdapter() error {
	if cfg.HTTPServerAddress != "" {
		go HTTPServer(cfg.HTTPServerAddress)
	}
	return nil
}
