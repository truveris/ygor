// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"fmt"
	"net/http"
)

type minionsHTMLHandler struct {
	*Server
}

func (handler *minionsHTMLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, err := auth(r)
	if err != nil {
		errorHandler(w, "Authentication failed", err)
		return
	}

	minions, err := handler.Server.Minions.All()
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
