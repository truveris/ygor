// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This file contains all the tools to handle the minions registry.
//

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	errMinionNotFound = errors.New("minion not found")
)

// MinionsFile is a wrapper around your minions file, it abstracts the
// serialization of minions and keeps an in-memory cache to avoid frequent
// reads.
type MinionsFile struct {
	path    string
	cache   map[string]*Minion
	lastMod time.Time
}

// Minion represents a single minion in memory.
type Minion struct {
	Name     string
	QueueURL string
	UserID   string
	LastSeen time.Time
}

// GetLine generates a simple line for persistence, with trailing new-line.
func (minion *Minion) GetLine() string {
	return fmt.Sprintf("%s\t%s\t%s\t%d\n", minion.Name, minion.QueueURL,
		minion.UserID, minion.LastSeen.Unix())
}

// OpenMinionsFile creates and return a wrapper around the file-system storage
// for minions.
func OpenMinionsFile(path string) (*MinionsFile, error) {
	file := &MinionsFile{path: path}
	err := file.reload()
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Check if the minion file has been updated. It also returns false if we can't
// read the file.
func (file *MinionsFile) needsReload() bool {
	if file.path == ":memory:" {
		return false
	}

	si, err := os.Stat(file.path)
	if err != nil {
		return false
	}

	// First update or the file was modified after the last update.
	if file.lastMod.IsZero() || si.ModTime().After(file.lastMod) {
		file.lastMod = si.ModTime()
		return true
	}

	return false
}

// All returns all the minions store in cache and on disk.
func (file *MinionsFile) All() ([]Minion, error) {
	var minions []Minion

	if file.needsReload() {
		err := file.reload()
		if err != nil {
			return nil, err
		}
	}

	for _, minion := range file.cache {
		minions = append(minions, *minion)
	}

	return minions, nil
}

// Get returns a minion with its name.
func (file *MinionsFile) Get(name string) (*Minion, error) {
	if file.needsReload() {
		err := file.reload()
		if err != nil {
			return nil, err
		}
	}

	for _, minion := range file.cache {
		if minion.Name == name {
			return minion, nil
		}
	}

	return nil, errMinionNotFound
}

// GetByUserID returns the minion registed with the given UserId.
func (file *MinionsFile) GetByUserID(userID string) (*Minion, error) {
	if file.needsReload() {
		err := file.reload()
		if err != nil {
			return nil, err
		}
	}

	for _, minion := range file.cache {
		if minion.UserID == userID {
			return minion, nil
		}
	}

	return nil, errMinionNotFound
}

// Register a minion.
func (file *MinionsFile) Register(name, queueURL, userID string) error {
	err := file.Add(name, queueURL, userID, time.Now())
	if err != nil {
		return err
	}
	return nil
}

// Add a new minion to the list, erasing any other minion with the same name.
// You want to use RegisterMinion if you want to check for errors.
func (file *MinionsFile) Add(name, queueURL, userID string, lastSeen time.Time) error {
	if file.needsReload() {
		err := file.reload()
		if err != nil {
			return err
		}
	}

	minion := &Minion{}
	minion.Name = name
	minion.QueueURL = queueURL
	minion.UserID = userID
	minion.LastSeen = lastSeen
	file.cache[minion.Name] = minion

	return nil
}

// Delete removes a minion from the in-memory storage.  You still need to call
// Save() to make it permanent.
func (file *MinionsFile) Delete(name string) {
	delete(file.cache, name)
}

// Save all the minions to disk.
func (file *MinionsFile) Save() error {
	if file.path == ":memory:" {
		return nil
	}

	// Maybe an easier way is to use ioutil.WriteFile
	fp, err := os.OpenFile(file.path, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer fp.Close()

	if len(file.cache) == 0 {
		fp.WriteString("\n")
		return nil
	}

	for _, minion := range file.cache {
		fp.WriteString(minion.GetLine())
	}

	return nil
}

func (file *MinionsFile) reload() error {
	file.cache = make(map[string]*Minion)

	if file.path == ":memory:" {
		return nil
	}

	// It's acceptable for the file not to exist at this point, we just
	// need to create it. Attempting to create it at this points allows us
	// to know early on whether the filesystem allows us to do so.
	fp, err := os.Open(file.path)
	if err != nil {
		if os.IsNotExist(err) {
			fp, err = os.Create(file.path)
			if err != nil {
				return err
			}
			fp.Close()
			return nil
		}
		return err
	}
	defer fp.Close()

	br := bufio.NewReader(fp)

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)

		// Break appart name and value.
		tokens := strings.Split(line, "\t")
		if len(tokens) != 4 {
			return errors.New("minion line is missing parameters")
		}

		lastSeenSinceEpoch, err := strconv.ParseInt(tokens[3], 10, 0)
		if err != nil {
			return errors.New("minion line has an invalid timestamp")
		}

		lastSeen := time.Unix(lastSeenSinceEpoch, 0)

		file.Add(tokens[0], tokens[1], tokens[2], lastSeen)
	}

	return nil
}
