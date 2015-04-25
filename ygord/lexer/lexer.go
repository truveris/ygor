// Copyright 2014-2015, Truveris Inc. All Rights Reserved.
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

package lexer

import (
	"io"
	"strings"
)

// Lexer is the lexical analysis struct, it's a tokenizer with some
// understanding of "sentences", separated by semi-colons.
type Lexer struct {
	t *Tokenizer
}

// New allocates a new Lexer struct with the given Reader as main input.
func New(input io.Reader) *Lexer {
	l := &Lexer{}
	l.t = NewTokenizer(input)
	return l
}

// NextSentence reads through the input ont token at a time to produce a single
// sentence.
func (l *Lexer) NextSentence() ([]string, error) {
	var sentence []string
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

// LexerSplit iterates through NextSentence calls to return a list of all the
// sentences parsed out from 's'.
func Split(s string) ([][]string, error) {
	var sentences [][]string

	l := New(strings.NewReader(s))

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
