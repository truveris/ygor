// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

type StringSet map[string]bool

func (ss StringSet) Add(s string) {
	ss[s] = true
}

func (ss StringSet) Array() []string {
	list := make([]string, len(ss))
	i := 0

	for s, _ := range ss {
		list[i] = s
		i++
	}

	return list
}
