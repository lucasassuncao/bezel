// Package fuzzy matches a query against a text the way a picker does: the
// query's letters in order, anywhere, case-blind, scored so that runs, word
// starts and an early start rank first.
package fuzzy

import (
	"slices"
	"unicode"
)

// Scores for each matched letter and what it lands on.
const (
	perMatch    = 16
	runBonus    = 15 // the letter follows the one matched before it
	wordBonus   = 10 // it starts a word
	firstBonus  = 8  // it is the text's first letter
	gapStart    = 3  // a gap opens between two matches
	gapPerRune  = 1
	lengthShare = 32 // a long text loses a point per this many letters
)

// Match reports whether query's letters appear in text in order, case-blind,
// with a score (higher ranks first) and the rune indices they matched. An
// empty query matches everything with a score of 0.
func Match(query, text string) (score int, at []int, ok bool) {
	q := []rune(query)
	if len(q) == 0 {
		return 0, nil, true
	}
	t := []rune(text)
	for i := range q {
		q[i] = unicode.ToLower(q[i])
	}
	end := forward(q, t)
	if end < 0 {
		return 0, nil, false
	}
	at = backward(q, t, end)
	return scoreOf(t, at), at, true
}

// forward is the index of the rune that completes the first in-order match
// scanning from the left, -1 when there is none.
func forward(q, t []rune) int {
	j := 0
	for i, r := range t {
		if unicode.ToLower(r) == q[j] {
			if j++; j == len(q) {
				return i
			}
		}
	}
	return -1
}

// backward walks back from end matching the query from its last letter, so
// the match is the tightest one ending there.
func backward(q, t []rune, end int) []int {
	at := make([]int, len(q))
	j := len(q) - 1
	for i := end; i >= 0 && j >= 0; i-- {
		if unicode.ToLower(t[i]) == q[j] {
			at[j] = i
			j--
		}
	}
	return at
}

func scoreOf(t []rune, at []int) int {
	s := 0
	for k, i := range at {
		s += perMatch
		switch {
		case i == 0:
			s += firstBonus + wordBonus
		case wordStart(t, i):
			s += wordBonus
		}
		if k > 0 {
			if gap := i - at[k-1] - 1; gap == 0 {
				s += runBonus
			} else {
				s -= gapStart + gap*gapPerRune
			}
		}
	}
	return s - at[0] - len(t)/lengthShare
}

// wordStart reports whether rune i begins a word: after a separator, or an
// upper-case letter after a lower-case one.
func wordStart(t []rune, i int) bool {
	prev, r := t[i-1], t[i]
	if !unicode.IsLetter(prev) && !unicode.IsDigit(prev) {
		return true
	}
	return unicode.IsLower(prev) && unicode.IsUpper(r)
}

// Ranked is the indices of texts that match query, best first; equal scores
// keep their order.
func Ranked(query string, texts []string) []int {
	type hit struct{ i, score int }
	var hits []hit
	for i, t := range texts {
		if s, _, ok := Match(query, t); ok {
			hits = append(hits, hit{i, s})
		}
	}
	slices.SortStableFunc(hits, func(a, b hit) int { return b.score - a.score })
	out := make([]int, len(hits))
	for k, h := range hits {
		out[k] = h.i
	}
	return out
}
