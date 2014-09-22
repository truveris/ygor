// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// This is ygord's lexical analyser, use the Split() function to transform this
// string:
//
//    unalias foo; alias foo "play http://foo/music.mp3; image bar.jpg"
//
// And creates a list of sentences, each containing a list of words, for this
// example, the output would be:
//
//    - ["unalias", "foo"]
//    - ["alias", "foo", "play http://foo/music.mp3; image bar.jpg"]
//
// Sitting at the highest level, the lexer leverages a tokenizer which, given
// the same stream returns:
//
//    - ["unalias", "foo", ";", "alias", "foo", "play http://foo/music.mp3; image bar.jpg"]
//
// The lexer essentially gets the stream of tokens/words and transform them
// into sentences, each separated by semi-colons.
//

package main

import (
	"io"
	"strings"
)

type Lexer struct {
	t *Tokenizer
}

func NewLexer(input io.Reader) *Lexer {
	l := &Lexer{}
	l.t = NewTokenizer(input)
	return l
}

func (l *Lexer) NextSentence() ([]string, error){
	sentence := make([]string, 0)
	var err error
	var word string

	for {
		word, err = l.t.NextToken()
		if err != nil && err != io.EOF {
			return nil, err
		}

		// End of sentence
		if word == ";" {
			break
		}

		sentence = append(sentence, word)

		if err == io.EOF {
			break
		}
	}

	return sentence, err
}

func LexerSplit(s string) ([][]string, error) {
	l := NewLexer(strings.NewReader(s))
	sentences := make([][]string, 0)

	for {
		sentence, err := l.NextSentence()
		if err != nil && err != io.EOF {
			return nil, err
		}

		if sentence != nil {
			sentences = append(sentences, sentence)
		}

		if err == io.EOF {
			return sentences, nil
		}
	}

	return sentences, nil
}
