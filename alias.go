// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This file contains all the tools to handle the aliases registry.
//

package ygor

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// Wrapper around your alias file, it abstracts the serialization of aliases
// and keeps an in-memory cache to avoid frequent reads.
type AliasFile struct {
	path string
	cache    map[string]*Alias
	lastMod  time.Time
}

// Definition of a single alias, used for in-memory storage.
type Alias struct {
	Name  string
	Value string
}

const (
	MaxRecursionLevel = 8
)

// Generate a simple line for persistence, with new-line.
func (alias *Alias) GetLine() string {
	return fmt.Sprintf("%s\t%s\n", alias.Name, alias.Value)
}

func (alias *Alias) SplitValue() (string, []string) {
	tokens := strings.Split(alias.Value, " ")
	return tokens[0], tokens[1:]
}

// Create and return a wrapper around the file-system storage for aliases.
func OpenAliasFile(path string) (*AliasFile, error) {
	file := &AliasFile{path: path}
	err := file.reload()
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Check if the underlying file has been updated. It also returns false if we
// can't read the file. XXX should return error instead.
func (file *AliasFile) needsReload() bool {
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

func (file *AliasFile) Get(name string) *Alias {
	if file.needsReload() {
		file.reload()
	}

	for _, alias := range file.cache {
		if alias.Name == name {
			return alias
		}
	}

	return nil
}

// Return a []string of all the alias names. FIXME: is there really no better
// way to get the keys of a map?
func (file *AliasFile) Names() []string {
	idx := 0
	names := make([]string, len(file.cache))
	for name, _ := range file.cache {
		names[idx] = name
		idx++
	}
	return names
}

func (file *AliasFile) Add(name, value string) {
	alias := &Alias{}
	alias.Name = name
	alias.Value = value
	file.cache[alias.Name] = alias
}

func (file *AliasFile) Delete(name string) {
	delete(file.cache, name)
}

// Save all the aliases to disk.
func (file *AliasFile) Save() error {
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

	for _, alias := range file.cache {
		fp.WriteString(alias.GetLine())
	}

	return nil
}

// Reload all the cached aliases from disk.
func (file *AliasFile) reload() error {
	file.cache = make(map[string]*Alias)

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
		} else {
			return err
		}
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
		tokens := strings.SplitN(line, "\t", 2)
		if len(tokens) != 2 {
			continue
		}

		file.Add(tokens[0], tokens[1])
	}

	return nil
}

// Recursively resolve aliases from a given line.
func (file *AliasFile) RecursiveResolve(line string, level int) (string, error) {
	if level >= MaxRecursionLevel {
		return line, errors.New("max recursion reached")
	}

	parts := strings.SplitN(line, " ", 2)

	// No more aliases, we're done here.
	alias := file.Get(parts[0])
	if alias == nil {
		return line, nil
	}

	// Build a new line from the alias.
	newparts := make([]string, 0)
	newparts = append(newparts, alias.Value)
	newparts = append(newparts, parts[1:]...)
	line = strings.Join(newparts, " ")

	line, err := file.RecursiveResolve(line, level+1)
	if err != nil {
		return "", err
	}

	return line, nil
}

// Recursively resolve aliases from a given line. Error out if we're 8 level
// deep and can't seem to resolve anything.
func (file *AliasFile) Resolve(line string) (string, error) {
	return file.RecursiveResolve(line, 0)
}
