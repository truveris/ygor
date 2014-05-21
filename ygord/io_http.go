// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// The io_irc_http is a temporary hack to expose ygor functions to the
// interwebs. This is bound to be replaced by an external process communicating
// with ygord via the SQS "API".
//

package main

import (
	"fmt"
	"log"
	"net/http"
)

func aliasesHandler(w http.ResponseWriter, r *http.Request) {
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
			<h1>ygor - aliases</h1>
			<table>
				<thead>
					<tr>
						<th>Name</th>
						<th>Value</th>
					</tr>
				</thead>
				<tbody>
	`)

	for _, alias := range aliases {
		fmt.Fprintf(w, `
		<tr>
			<td>%s</td>
			<td>%s</td>
		</tr>`, alias.Name, alias.Value)
	}

	fmt.Fprintf(w, `
		</body>
	</html>
	`)
}

func minionsHandler(w http.ResponseWriter, r *http.Request) {
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
			<h2>ygor - minions</h2>
			<table>
				<thead>
					<tr>
						<th>Name</th>
						<th>Last Seen</th>
					</tr>
				</thead>
				<tbody>
	`)

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

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
	<html>
	<head><title>ygor</title></head>
	<body>
		<h1>ygor</h1>
		<ul>
			<li><a href="/aliases">aliases</a>
			<li><a href="/minions">minions</a>
		</ul>
	</body>
	</html>
	`)
}

func HTTPServer(address string) {
	log.Printf("starting http server on %s", address)
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/aliases", aliasesHandler)
	http.HandleFunc("/minions", minionsHandler)
	http.ListenAndServe(address, nil)
}

func StartHTTPAdapter() error {
	if cfg.HTTPServerAddress != "" {
		go HTTPServer(cfg.HTTPServerAddress)
	}
	return nil
}
