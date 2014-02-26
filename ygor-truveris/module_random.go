// Copyright 2014, Truveris Inc. All Rights Reserved.
// Use of this source code is governed by the ISC license in the LICENSE file.
//
// Come up with random things to say on timer or when someone talks to the bot.
//

package main

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

var (
	MaxWordsPerSentence = 50
)

type RandomModule struct {
	Markov      *Chain
	LastChannel string
}

func (module RandomModule) MakeSentence() string {
	return module.Markov.Generate(MaxWordsPerSentence)
}

// The topic is a two words string that is used as the bootstrap for a
// sentence.
func (module RandomModule) MakeSentenceOnTopic(topic string) string {
	var output string
	maxWords := MaxWordsPerSentence / 2

	output = module.Markov.GenerateBackward(topic, maxWords) + " " + topic

	// If the topic is the end of the sentence, leave it there.
	if !strings.HasSuffix(topic, ".") {
		output = output + " " + module.Markov.GenerateForward(topic,
			maxWords)
	}

	return output
}

// Given a PRIVMSG, return the list of two word topics order with the least
// common first (less likely to be common words).
func (module RandomModule) sortedTopics(msg string) []string {
	words := &ScoredWords{}

	// Loop through the msg, by tuples of two. Find the tuple with the
	// lowest amount of match within the chains.
	var a, b string
	for _, word := range strings.Fields(msg) {
		a, b = b, word
		score := 0
		key := a + " " + b
		keyWords, ok := module.Markov.forward[key]
		if ok {
			score = len(keyWords)
		}
		keyWords, ok = module.Markov.backward[key]
		if ok {
			score = score + len(keyWords)
		}

		if score > 0 {
			words.Append(key, score)
		}
	}

	sort.Sort(words)

	return words.Strings()
}

// Send a coherent sentence to the markov chain. Strip the prefix nickname
// addresses, use the nickname instead of /ME or action.
func (module RandomModule) recordPrivMsg(nick, msg string, isAction bool) {
	var sentence string

	if isAction {
		sentence = nick + " " + msg
	} else {
		sentence = msg
	}

	// We don't want to record the nickname before sentences, strip
	// anything before a colon.
	sentence = strings.Split(sentence, ":")[0]
	sentence = strings.TrimSpace(sentence)

	module.Markov.AddLine(sentence)

	// Save sentence to disk.
	flags := os.O_APPEND | os.O_CREATE
	file, err := os.OpenFile("data/truveris/learning", flags, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[soul] %s\n", err.Error())
		return
	}

	fmt.Fprintf(file, "%s\n", sentence)
	file.Close()
}

// React to PRIVMSG (channel or direct).
func (module RandomModule) PrivMsg(nick, where, msg string, isAction bool) {
	if nick != cmd.Nickname {
		module.LastChannel = where
	}

	module.recordPrivMsg(nick, msg, isAction)

	// We only care for messages addressed to us.
	if !strings.HasPrefix(msg, cmd.Nickname) {
		return
	}

	topics := module.sortedTopics(msg)

	var text string
	if len(topics) > 0 {
		text = module.MakeSentenceOnTopic(topics[0])
	} else {
		text = module.MakeSentence()
	}

	privMsg(where, text)
}

// Every 10 seconds, decide whether we have something clever to say.
func (module RandomModule) Ticker() {
	ticker := time.Tick(10 * time.Second)
	chances := 1.0 / 1000.0

	for _ = range ticker {
		if rand.Float64() < chances {
			text := module.MakeSentence()
			privMsg(module.LastChannel, text)
		}
	}
}

// Initialize the Random module: start the ticker.
func (module RandomModule) Init() {
	go module.Ticker()
}
