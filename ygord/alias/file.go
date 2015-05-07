// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This file contains all the tools to handle the aliases registry.  The
// reserved file path ":memory:" will cause this implementation to never access
// the file-system and always start from a blank slate.
//

package alias

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/truveris/ygor/ygord/lexer"
)

var (
	errTooManyHashes  = errors.New("too many '#'")
	errTooManyAliases = errors.New("too many aliases with this prefix")
)

const (
	// MaxRecursionLevel defines the number or run allowed while resolving
	// an alias.
	MaxRecursion = 8

	// MaxAliasIncrements defines how far we should check for increments
	// within the same namespace.  This is really just here to avoid abuse
	// more than anything.
	MaxAliasIncrements = 10000
)

// File wraps your alias file, it abstracts the serialization of aliases and
// keeps an in-memory cache to avoid frequent reads.
type File struct {
	path    string
	cache   map[string]*Alias
	lastMod time.Time
}

// Open creates and returns a wrapper around the file-system storage for
// aliases.
func Open(path string) (*File, error) {
	file := &File{path: path}
	err := file.reload()
	if err != nil {
		return nil, err
	}
	return file, nil
}

// Check if the underlying file has been updated. It also returns false if we
// can't read the file. XXX should return error instead.
func (file *File) needsReload() bool {
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

// Get returns the alias given its name.  Returns nil if not found.
func (file *File) Get(name string) *Alias {
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

// Names returns a sorted list of all the alias names.
func (file *File) Names() []string {
	idx := 0
	names := make([]string, len(file.cache))
	for name := range file.cache {
		names[idx] = name
		idx++
	}
	sort.Strings(names)
	return names
}

// Add creates a new alias in the in-memory cache.  It will be saved
// permanently once Save is called.
func (file *File) Add(name, value, author string, time time.Time) {
	alias := &Alias{}
	alias.Name = name
	alias.Value = value
	alias.Author = author
	alias.CreationTime = time
	file.cache[alias.Name] = alias
}

// Delete removes an alias by name from the local cache. It will not be saved
// permanently until Save is called.
func (file *File) Delete(name string) {
	delete(file.cache, name)
}

// Save all the aliases to disk.
func (file *File) Save() error {
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

	for _, alias := range file.cache {
		fp.WriteString(alias.String() + "\n")
	}

	return nil
}

// Reload all the cached aliases from disk.
func (file *File) reload() error {
	file.cache = make(map[string]*Alias)

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
		tokens := strings.SplitN(line, "\t", 4)
		if len(tokens) != 4 {
			continue
		}

		date, err := time.Parse(time.RFC3339, tokens[3])
		if err != nil {
			date = time.Now()
		}

		file.Add(tokens[0], tokens[1], tokens[2], date)
	}

	return nil
}

// Resolve recursively resolves aliases from a given line. Error out if the
// MaxRecursionLevel is reached and we're not getting anywhere.
func (file *File) Resolve(line string, recursion int) (string, error) {
	if recursion >= MaxRecursion {
		return line, errors.New("max recursion reached")
	}

	// Only resolve the first word of a line.
	parts := strings.SplitN(line, " ", 2)

	// No more aliases, we're done here.
	alias := file.Get(parts[0])
	if alias == nil {
		return line, nil
	}

	// Build a new line from the alias.
	if len(parts) > 1 {
		line = alias.Value + " " + parts[1]
	} else {
		line = alias.Value
	}

	line, err := file.Resolve(line, recursion+1)
	if err != nil {
		return "", err
	}

	return line, nil
}

// All returns all the aliases in the system.
func (file *File) All() ([]Alias, error) {
	var aliases []Alias

	if file.needsReload() {
		err := file.reload()
		if err != nil {
			return nil, err
		}
	}

	names := file.Names()

	for _, name := range names {
		alias := *file.cache[name]
		aliases = append(aliases, alias)
	}

	return aliases, nil
}

// Find returns the name of all aliases matching the provided pattern.
func (file *File) Find(pattern string) []string {
	var results []string

	aliases := file.Names()

	for _, name := range aliases {
		if strings.Contains(name, pattern) {
			results = append(results, name)
		}
	}

	return results
}

func (file *File) ExpandSentence(words []string, recursion int) ([][]string, error) {
	sentences := make([][]string, len(words))

	// Resolve any alias found as first word.
	expanded, err := file.Resolve(words[0], recursion)
	if err != nil {
		return nil, err
	}

	sentences, err = lexer.Split(expanded)
	if err != nil {
		return nil, err
	}

	last := len(sentences) - 1
	sentences[last] = append(sentences[last], words[1:]...)

	return sentences, nil
}

// Expand sentences through aliases.
func (file *File) ExpandSentences(ss [][]string, recursion int) ([][]string, error) {
	sentences := make([][]string, len(ss))

	for _, words := range ss {
		if len(words) == 0 {
			continue
		}

		newsentences, err := file.ExpandSentence(words, recursion)
		if err != nil {
			return nil, err
		}
		sentences = append(sentences, newsentences...)
	}

	return sentences, nil
}

// GetIncrementedName returns the next available alias with a trailing number
// incremented if needed.  This is used when an alias has a trailing '#'.
func (file *File) GetIncrementedName(name, value string) (string, error) {
	cnt := strings.Count(name, "#")
	if cnt == 0 {
		return name, nil
	} else if cnt > 1 {
		return "", errTooManyHashes
	}

	var newName string

	for i := 1; ; i++ {
		newName = strings.Replace(name, "#", fmt.Sprintf("%d", i), 1)

		if i > MaxAliasIncrements {
			return "", errTooManyAliases
		}

		alias := file.Get(newName)
		if alias == nil {
			break
		}

		if alias.Value == value {
			return "", errors.New("already exists as '" + alias.Name + "'")
		}
	}

	return newName, nil
}
