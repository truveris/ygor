// Copyright 2014-2015, Truveris Inc. All Rights Reserved.

package main

// Ping replies to ygord with the same timestamp, helping it estimate the
// duration of a round-trip to this minion.
func Ping(timestamp string) {
	Send("pong " + timestamp)
}
