// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This file contains all the tools to handle the aliases registry.
//

package alias

import (
	"fmt"
	"strings"
	"time"
)

// Alias is the definition of a single alias, used for in-memory storage.
type Alias struct {
	Name         string
	Value        string
	Author       string
	CreationTime time.Time
}

// GetLine generates a single line for writing on file.
func (alias *Alias) String() string {
	creationTime := alias.CreationTime.Format(time.RFC3339)
	return fmt.Sprintf("%s\t%s\t%s\t%s", alias.Name, alias.Value,
		alias.Author, creationTime)
}

// SplitValue just splits the alias value into the first part (likely the
// command) and everything else (the arguments).
func (alias *Alias) SplitValue() (string, []string) {
	tokens := strings.Split(alias.Value, " ")
	return tokens[0], tokens[1:]
}

// HumanTime returns a pretty version of the time for display.
func (alias *Alias) HumanTime() string {
	return alias.CreationTime.Format(time.RFC3339)
}
