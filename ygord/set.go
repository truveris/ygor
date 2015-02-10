// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

// StringSet is a string->bool map used a set of string.
type StringSet map[string]bool

// Add an element to the set.
func (ss StringSet) Add(s string) {
	ss[s] = true
}

// Array returns an array version of this set.
func (ss StringSet) Array() []string {
	list := make([]string, len(ss))
	i := 0

	for s := range ss {
		list[i] = s
		i++
	}

	return list
}
