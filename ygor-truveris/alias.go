// Copyright (c) 2014 Bertrand Janin <b@truveris.com>
// Use of this source code is governed by the ISC license in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

type Alias struct {
	Name  string
	Value string
}

const (
	AliasFilename = "aliases.cfg"
)

var (
	Aliases = make(map[string]*Alias)
	LastMod time.Time
)

// Generate a simple line for persistence, with new-line.
func (alias *Alias) GetLine() string {
	return fmt.Sprintf("%s\t%s\n", alias.Name, alias.Value)
}

func (alias *Alias) SplitValue() (string, []string) {
	tokens := strings.Split(alias.Value, " ")
	return tokens[0], tokens[1:]
}

// Check if the alias file has been updated. It also returns false if we can't
// read the file.
func aliasesNeedReload() bool {
	si, err := os.Stat(AliasFilename)
	if err != nil {
		return false
	}

	// First update or the file was modified after the last update.
	if LastMod.IsZero() || si.ModTime().After(LastMod) {
		LastMod = si.ModTime()
		return true
	}

	return false
}

func GetAlias(name string) *Alias {
	if aliasesNeedReload() {
		reloadAliases()
	}

	for _, alias := range Aliases {
		if alias.Name == name {
			return alias
		}
	}

	return nil
}

func AddAlias(name, value string) {
	alias := &Alias{}
	alias.Name = name
	alias.Value = value
	Aliases[alias.Name] = alias
}

func DeleteAlias(name) {
	delete(Aliases, name)
}

// Save all the aliases to disk.
func SaveAliases() {
	file, err := os.OpenFile(AliasFilename, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		// TODO debug channel
		return
	}
	defer file.Close()

	for _, alias := range Aliases {
		file.Write([]byte(alias.GetLine()))
	}

}

func reloadAliases() {
	Aliases = make(map[string]*Alias)

	file, err := os.Open(AliasFilename)
	if err != nil {
		return
	}

	br := bufio.NewReader(file)

	for {
		line, err := br.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)

		// Break appart name and value.
		tokens := strings.SplitN(line, "\t", 2)
		if len(tokens) != 2 {
			continue
		}

		AddAlias(tokens[0], tokens[1])
	}

	// XXX send that to debug channel
	// fmt.Printf("(Re-)loaded %d aliases.\n", len(Aliases))
}
