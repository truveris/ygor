// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"html/template"
	"net/http"
)

var (
	rootTmplRaw = `
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
		<h1>{{.Username}}@ygor</h1>
		<ul>
			<li><a href="/aliases">aliases</a>
			<li><a href="/minions">minions</a>
			<li><a href="/channels">channels</a>
		</ul>
	</body>
	</html>`
	rootTmpl = template.Must(template.New("root").Parse(rootTmplRaw))
)

type rootHTMLContext struct {
	Username string
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth(r)
	if err != nil {
		errorHandler(w, "Authentication failed", err)
		return
	}

	ctx := rootHTMLContext{Username: user}

	err = rootTmpl.Execute(w, ctx)
	if err != nil {
		http.Error(w, "error: "+err.Error(), 500)
		return
	}
}
