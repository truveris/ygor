// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This is ygord's tokenizer. It takes strings of characters such as:
//
//    unalias foo;alias foo "play http://foo/music.mp3; image bar.jpg"
//
// And creates a list of tokens:
//
//    - ["unalias", "foo", ";", "alias", "foo", "play http://foo/music.mp3; image bar.jpg"]
//
// Note that unescaped semi-colons are singled out as tokens in a shell
// fashion. They are used to delimite statements/sentences at the lexer level.
//
// This tokenizer has some peculiarities:
//    - \a-z will remain unescaped (\a remains \a).
//    - \\ should be used to produce a single \
//    - \" should be used to produce a single "
//

package main

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

// Tokenizer States
const (
	StateStart         TokenizerState = iota
	StateInWord        TokenizerState = iota
	StateQuoted        TokenizerState = iota
	StateQuotedEscaped TokenizerState = iota
	StateEscaped       TokenizerState = iota
	StateInWordEscaped TokenizerState = iota
)

// Rune types
const (
	RuneTypeEOF       RuneType = iota
	RuneTypeEOS       RuneType = iota
	RuneTypeChar      RuneType = iota
	RuneTypeSpace     RuneType = iota
	RuneTypeQuote     RuneType = iota
	RuneTypeBackslash RuneType = iota
)

// RuneType defines a rune type as defined in the constants above.
type RuneType int

// TokenizerState defines the current state of the Tokenizer, should be one the
// constants above with the State* prefix.
type TokenizerState int

// Tokenizer wraps a Reader from which tokens are extracted.
type Tokenizer struct {
	input *bufio.Reader
}

// NewTokenizer allocates a new Tokenizer from the provided Reader.
func NewTokenizer(input io.Reader) *Tokenizer {
	t := &Tokenizer{bufio.NewReader(input)}
	return t
}

// GetRuneType returns the type of the given run.
func (t *Tokenizer) GetRuneType(r rune) RuneType {
	switch {
	case r == ';':
		return RuneTypeEOS
	case r == '"':
		return RuneTypeQuote
	case r == '\\':
		return RuneTypeBackslash
	case strings.ContainsRune(" \t\n\r", r):
		return RuneTypeSpace
	}

	return RuneTypeChar
}

// NextToken reads a single token from the tokenizer's input.
func (t *Tokenizer) NextToken() (string, error) {
	state := StateStart
	var token []rune
	for {
		nextRune, _, err := t.input.ReadRune()
		nextRuneType := t.GetRuneType(nextRune)
		if err != nil {
			if err == io.EOF {
				nextRuneType = RuneTypeEOF
			} else {
				return "", err
			}
		}

		switch state {
		case StateStart:
			switch nextRuneType {
			case RuneTypeEOF:
				return "", io.EOF
			case RuneTypeChar:
				state = StateInWord
				token = append(token, nextRune)
			case RuneTypeEOS:
				return ";", nil
			case RuneTypeSpace:
				continue
			case RuneTypeBackslash:
				state = StateInWordEscaped
			case RuneTypeQuote:
				state = StateQuoted
			}
		case StateInWord:
			switch nextRuneType {
			case RuneTypeEOF:
				return string(token), io.EOF
			case RuneTypeChar:
				token = append(token, nextRune)
			case RuneTypeSpace:
				return string(token), nil
			case RuneTypeBackslash:
				state = StateInWordEscaped
			case RuneTypeEOS:
				t.input.UnreadRune()
				return string(token), nil
			case RuneTypeQuote:
				state = StateQuoted
			}
		case StateQuoted:
			switch nextRuneType {
			case RuneTypeEOF:
				return string(token), errors.New("missing quote termination")
			case RuneTypeChar, RuneTypeSpace, RuneTypeEOS:
				token = append(token, nextRune)
			case RuneTypeBackslash:
				state = StateQuotedEscaped
			case RuneTypeQuote:
				state = StateInWord
			}
		case StateInWordEscaped:
			switch nextRuneType {
			case RuneTypeEOF:
				return string(token), errors.New("unterminated escape character")
			case RuneTypeChar:
				token = append(token, '\\')
				token = append(token, nextRune)
				state = StateInWord
			case RuneTypeQuote:
				token = append(token, '"')
				state = StateInWord
			case RuneTypeBackslash:
				token = append(token, '\\')
				state = StateInWord
			default:
				return string(token), errors.New("unknown escape character")
			}
		case StateQuotedEscaped:
			switch nextRuneType {
			case RuneTypeEOF:
				return string(token), errors.New("unterminated escape character")
			case RuneTypeChar:
				token = append(token, '\\')
				token = append(token, nextRune)
				state = StateQuoted
			case RuneTypeQuote:
				token = append(token, '"')
				state = StateQuoted
			case RuneTypeBackslash:
				token = append(token, '\\')
				state = StateQuoted
			default:
				return string(token), errors.New("unknown escape character")
			}
		}
	}
	return "", nil
}
