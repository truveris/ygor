// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This file contains all the tools to handle the minions registry.
//

package ygor

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Minion struct {
	Name     string
	QueueURL string
	UserID   string
	LastSeen time.Time
}

var (
	// This is a default value, it can be changed with the
	// SetMinionFilePath function.
	minionsFilePath = "minions.cfg"

	// TODO: make this private at some point...
	Minions        = make(map[string]*Minion)
	MinionsLastMod time.Time
)

// Generate a simple line for persistence, with new-line.
func (minion *Minion) GetLine() string {
	return fmt.Sprintf("%s\t%s\t%s\t%d\n", minion.Name, minion.QueueURL,
		minion.UserID, minion.LastSeen.Unix())
}

func (minion *Minion) SplitValue() (string, []string) {
	tokens := strings.Split(minion.QueueURL, " ")
	return tokens[0], tokens[1:]
}

// Check if the minion file has been updated. It also returns false if we can't
// read the file.
func MinionsNeedReload() bool {
	si, err := os.Stat(minionsFilePath)
	if err != nil {
		return false
	}

	// First update or the file was modified after the last update.
	if MinionsLastMod.IsZero() || si.ModTime().After(MinionsLastMod) {
		MinionsLastMod = si.ModTime()
		return true
	}

	return false
}

func GetMinions() ([]Minion, error) {
	minions := make([]Minion, 0)
	if MinionsNeedReload() {
		err := ReloadMinions()
		if err != nil {
			return nil, err
		}
	}

	for _, minion := range Minions {
		minions = append(minions, *minion)
	}

	return minions, nil
}

func GetMinion(name string) (*Minion, error) {
	if MinionsNeedReload() {
		err := ReloadMinions()
		if err != nil {
			return nil, err
		}
	}

	for _, minion := range Minions {
		if minion.Name == name {
			return minion, nil
		}
	}

	return nil, errors.New("minion not found")
}

func GetMinionByUserID(userID string) (*Minion, error) {
	if MinionsNeedReload() {
		err := ReloadMinions()
		if err != nil {
			return nil, err
		}
	}

	for _, minion := range Minions {
		if minion.UserID == userID {
			return minion, nil
		}
	}

	return nil, errors.New("minion not found")
}

// Register a minion.
func RegisterMinion(name, queueURL, userID string) error {
	err := AddMinion(name, queueURL, userID, time.Now())
	if err != nil {
		return err
	}
	return nil
}

// Add a new minion to the list, erasing any other minion with the same name.
// You want to use RegisterMinion if you want to check for errors.
func AddMinion(name, queueURL, userID string, lastSeen time.Time) error {
	if MinionsNeedReload() {
		err := ReloadMinions()
		if err != nil {
			return err
		}
	}

	minion := &Minion{}
	minion.Name = name
	minion.QueueURL = queueURL
	minion.UserID = userID
	minion.LastSeen = lastSeen
	Minions[minion.Name] = minion

	return nil
}

func DeleteMinion(name string) {
	delete(Minions, name)
}

// Save all the minions to disk.
func SaveMinions() error {
	// Maybe an easier way is to use ioutil.WriteFile
	file, err := os.OpenFile(minionsFilePath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	if len(Minions) == 0 {
		file.WriteString("\n")
		return nil
	}

	for _, minion := range Minions {
		file.WriteString(minion.GetLine())
	}

	return nil
}

func ReloadMinions() error {
	Minions = make(map[string]*Minion)

	file, err := os.Open(minionsFilePath)
	if err != nil {
		return err
	}

	br := bufio.NewReader(file)

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

		AddMinion(tokens[0], tokens[1], tokens[2], lastSeen)
	}

	return nil
}

func SetMinionsFilePath(path string) {
	minionsFilePath = path
}
