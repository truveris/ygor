// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This file contains all the tools to handle the aliases registry.
//

package alias

import (
	"fmt"
	"strings"
)

// Alias is the definition of a single alias, used for in-memory storage.
type Alias struct {
	Name  string
	Value string
}

// GetLine generates a single line for writing on file.
func (alias *Alias) GetLine() string {
	return fmt.Sprintf("%s\t%s\n", alias.Name, alias.Value)
}

// SplitValue just splits the alias value into the first part (likely the
// command) and everything else (the arguments).
func (alias *Alias) SplitValue() (string, []string) {
	tokens := strings.Split(alias.Value, " ")
	return tokens[0], tokens[1:]
}
