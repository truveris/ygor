// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package lexer

import (
	"strings"
	"testing"
)

func EqualSentences(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func AssertNextSentence(l *Lexer, t *testing.T, a []string) {
	b, _ := l.NextSentence()

	if !EqualSentences(a, b) {
		t.Errorf("bad sentence: '%s' VS '%s'", a, b)
	}
}

func TestLexer(t *testing.T) {
	s := strings.NewReader(`foo bar; spaces  "and quotes"; "es\"ca\\p e"`)
	z := NewLexer(s)

	AssertNextSentence(z, t, []string{"foo", "bar"})
	AssertNextSentence(z, t, []string{"spaces", "and quotes"})
	AssertNextSentence(z, t, []string{`es"ca\p e`})
}

func TestLexerSplit(t *testing.T) {
	ss, err := LexerSplit("foo bar baz")
	if err != nil {
		t.Error(err)
	}

	if len(ss) != 1 {
		t.Errorf("bad result: %s", ss)
	}
	if !EqualSentences(ss[0], []string{"foo", "bar", "baz"}) {
		t.Errorf("bad result: %s", ss)
	}
}
