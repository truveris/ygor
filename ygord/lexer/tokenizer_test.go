// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.

package lexer

import (
	"strings"
	"testing"
)

func AssertNextToken(z *Tokenizer, t *testing.T, s string) {
	token, _ := z.NextToken()
	if token != s {
		t.Errorf("bad token: '%s' != '%s'", token, s)
	}
}

func TestTokenizerSpace(t *testing.T) {
	s := strings.NewReader(`foo bar f00`)
	z := NewTokenizer(s)

	AssertNextToken(z, t, "foo")
	AssertNextToken(z, t, "bar")
	AssertNextToken(z, t, "f00")
}

func TestTokenizerSemiColon(t *testing.T) {
	s := strings.NewReader(`foo;bar;f00`)
	z := NewTokenizer(s)

	AssertNextToken(z, t, "foo")
	AssertNextToken(z, t, ";")
	AssertNextToken(z, t, "bar")
	AssertNextToken(z, t, ";")
	AssertNextToken(z, t, "f00")
}

func TestTokenizerQuoted(t *testing.T) {
	s := strings.NewReader(`foo "bar f00"`)
	z := NewTokenizer(s)

	AssertNextToken(z, t, "foo")
	AssertNextToken(z, t, "bar f00")
}

func TestTokenizerQuotedQuote(t *testing.T) {
	s := strings.NewReader(`foo "bar \"f00\""`)
	z := NewTokenizer(s)

	AssertNextToken(z, t, "foo")
	AssertNextToken(z, t, `bar "f00"`)
}

func TestTokenizerQuotedBackslash(t *testing.T) {
	s := strings.NewReader(`foo "bar\\f00"`)
	z := NewTokenizer(s)

	AssertNextToken(z, t, "foo")
	AssertNextToken(z, t, `bar\f00`)
}
