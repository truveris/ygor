// Copyright (c) 2014 Bertrand Janin <b@janin.com>
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// List of words with scores. This is compatible with the sort.Sort function.

package main

type ScoredWord struct {
	word  string
	score int
}

// Sortable list of scored words.
type ScoredWords struct {
	words []ScoredWord
}

// Number of sorted words (sort.Sort data interface)
func (words *ScoredWords) Len() int {
	return len(words.words)
}

// Compare two sorted words (sort.Sort data interface)
func (words *ScoredWords) Less(i, j int) bool {
	if words.words[i].score < words.words[j].score {
		return true
	}
	return false
}

// Swap two sorted words (sort.Sort data interface)
func (words *ScoredWords) Swap(i, j int) {
	temp := words.words[i]
	words.words[i] = words.words[j]
	words.words[j] = temp
}

// Add one sorted words to the structure.
func (words *ScoredWords) Append(key string, score int) {
	words.words = append(words.words, ScoredWord{key, score})
}

// Return the array of strings without their scores, in the same order.
func (words *ScoredWords) Strings() []string {
	strings := make([]string, words.Len())
	for i, word := range words.words {
		strings[i] = word.word
	}
	return strings
}
