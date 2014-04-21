// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"testing"
)

func TestGetReversedArrayThree(t *testing.T) {
	input := []string{"a", "b", "c"}
	output := getReversedArray(input)

	if len(output) != len(input) {
		t.Fatal("wrong length (%d)", len(output))
	}

	if output[0] != "c" || output[1] != "b" || output[2] != "a" {
		t.Fatal("not reversed: %s", output)
	}
}

func TestGetReversedArrayOne(t *testing.T) {
	input := []string{"a"}
	output := getReversedArray(input)

	if len(output) != len(input) {
		t.Fatal("wrong length (%d)", len(output))
	}

	if output[0] != "a" {
		t.Fatal("not reversed: %s", output)
	}
}

func TestLeaderUnshift(t *testing.T) {
	input := make(Leader, 3)
	input[0] = "a"
	input[1] = "b"
	input[2] = "c"

	input.Unshift("Z")

	if len(input) != 3 {
		t.Fatal("wrong length (%d)", len(input))
	}

	if input[0] != "Z" || input[1] != "a" || input[2] != "b" {
		t.Fatal("not unshifted: %s", input)
	}
}
