// Copyright 2014, Truveris Inc. All Rights Reserved.
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

var (
)

const (
	STATE_START TokenizerState = iota
	STATE_INWORD TokenizerState = iota
	STATE_QUOTED TokenizerState = iota
	STATE_QUOTED_ESCAPED TokenizerState = iota
	STATE_ESCAPED TokenizerState = iota
	STATE_INWORD_ESCAPED TokenizerState = iota

	RUNETYPE_EOF RuneType = iota
	RUNETYPE_EOS RuneType = iota
	RUNETYPE_CHAR RuneType = iota
	RUNETYPE_SPACE RuneType = iota
	RUNETYPE_QUOTE RuneType = iota
	RUNETYPE_BACKSLASH RuneType = iota
)

type RuneType int
type TokenizerState int

type Tokenizer struct {
	input *bufio.Reader
}

func NewTokenizer(input io.Reader) *Tokenizer {
	t := &Tokenizer{bufio.NewReader(input)}
	return t
}

func (t *Tokenizer) GetRuneType(r rune) RuneType {
	switch {
	case r == ';':
		return RUNETYPE_EOS
	case r == '"':
		return RUNETYPE_QUOTE
	case r == '\\':
		return RUNETYPE_BACKSLASH
	case strings.ContainsRune(" \t\n\r", r):
		return RUNETYPE_SPACE
	}

	return RUNETYPE_CHAR
}

func (t *Tokenizer) NextToken() (string, error) {
	state := STATE_START
	token := make([]rune, 0)
	for {
		nextRune, _, err := t.input.ReadRune()
		nextRuneType := t.GetRuneType(nextRune)
		if err != nil {
			if err == io.EOF {
				nextRuneType = RUNETYPE_EOF
			} else {
				return "", err
			}
		}

		switch state {
		case STATE_START:
			switch nextRuneType {
			case RUNETYPE_EOF:
				return "", io.EOF
			case RUNETYPE_CHAR:
				state = STATE_INWORD
				token = append(token, nextRune)
			case RUNETYPE_EOS:
				return ";", nil
			case RUNETYPE_SPACE:
				continue
			case RUNETYPE_BACKSLASH:
				state = STATE_INWORD_ESCAPED
			case RUNETYPE_QUOTE:
				state = STATE_QUOTED
			}
		case STATE_INWORD:
			switch nextRuneType {
			case RUNETYPE_EOF:
				return string(token), io.EOF
			case RUNETYPE_CHAR:
				token = append(token, nextRune)
			case RUNETYPE_SPACE:
				return string(token), nil
			case RUNETYPE_BACKSLASH:
				state = STATE_INWORD_ESCAPED
			case RUNETYPE_EOS:
				t.input.UnreadRune()
				return string(token), nil
			case RUNETYPE_QUOTE:
				state = STATE_QUOTED
			}
		case STATE_QUOTED:
			switch nextRuneType {
			case RUNETYPE_EOF:
				return string(token), errors.New("missing quote termination")
			case RUNETYPE_CHAR, RUNETYPE_SPACE, RUNETYPE_EOS:
				token = append(token, nextRune)
			case RUNETYPE_BACKSLASH:
				state = STATE_QUOTED_ESCAPED
			case RUNETYPE_QUOTE:
				state = STATE_INWORD
			}
		case STATE_INWORD_ESCAPED:
			switch nextRuneType {
			case RUNETYPE_EOF:
				return string(token), errors.New("unterminated escape character")
			case RUNETYPE_CHAR:
				token = append(token, '\\')
				token = append(token, nextRune)
				state = STATE_INWORD
			case RUNETYPE_QUOTE:
				token = append(token, '"')
				state = STATE_INWORD
			case RUNETYPE_BACKSLASH:
				token = append(token, '\\')
				state = STATE_INWORD
			default:
				return string(token), errors.New("unknown escape character")
			}
		case STATE_QUOTED_ESCAPED:
			switch nextRuneType {
			case RUNETYPE_EOF:
				return string(token), errors.New("unterminated escape character")
			case RUNETYPE_CHAR:
				token = append(token, '\\')
				token = append(token, nextRune)
				state = STATE_QUOTED
			case RUNETYPE_QUOTE:
				token = append(token, '"')
				state = STATE_QUOTED
			case RUNETYPE_BACKSLASH:
				token = append(token, '\\')
				state = STATE_QUOTED
			default:
				return string(token), errors.New("unknown escape character")
			}
		}
	}
	return "", nil
}
