// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package alias

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestAliasResolve(t *testing.T) {
	expected := "image http://i.imgur.com/uVIqN.jpg appendage"

	f, err := ioutil.TempFile("", "ygor-test-")
	if err != nil {
		t.Error(err)
	}

	a, err := Open(f.Name())
	if err != nil {
		t.Error(err)
	}

	a.Add("noodly", "image http://i.imgur.com/uVIqN.jpg", "fsm", time.Now())
	a.Save()
	line, err := a.Resolve("noodly appendage")
	if err != nil {
		t.Error(err)
	}

	f.Close()
	os.Remove(f.Name())

	if line != expected {
		t.Errorf("output does not match: %s != %s", expected, line)
	}
}

func TestAliasResolveRecursive(t *testing.T) {
	expected := "web http://i.imgur.com/uVIqN.jpg appendage"

	f, err := ioutil.TempFile("", "ygor-test-")
	if err != nil {
		t.Error(err)
	}

	a, err := Open(f.Name())
	if err != nil {
		t.Error(err)
	}

	a.Add("image", "web", "fsm", time.Now())
	a.Add("noodly", "image http://i.imgur.com/uVIqN.jpg", "fsm", time.Now())
	a.Save()
	line, err := a.Resolve("noodly appendage")
	if err != nil {
		t.Error(err)
	}

	f.Close()
	os.Remove(f.Name())

	if line != expected {
		t.Errorf("output does not match: %s != %s", expected, line)
	}
}
