package hw03frequencyanalysis

import (
	"maps"
	"regexp"
	"slices"
	"sort"
	"strings"
)

const (
	Delimeter   = " "
	SelectCount = 10
)

type WordStatistic struct {
	Word  string
	Count int
}

var (
	regexFilter = regexp.MustCompile(`[^\p{Cyrillic}-]`)
	skipWords   = map[string]struct{}{"-": {}}
)

func Top10(input string) []string {
	if input == "" {
		return []string{}
	}

	words := strings.Fields(input)
	wordsStatistic := countWords(words, skipWords)

	sort.Slice(wordsStatistic, func(i, j int) bool { return wordsStatisticComparer(wordsStatistic[i], wordsStatistic[j]) })

	return firstWords(wordsStatistic, SelectCount)
}

func countWords(words []string, skip map[string]struct{}) []*WordStatistic {
	counter := make(map[string]*WordStatistic)

	for _, word := range words {
		wordWithoutPunctuation := regexFilter.ReplaceAllString(word, "")
		lowerCaseWord := strings.ToLower(wordWithoutPunctuation)

		if _, ok := skip[lowerCaseWord]; ok {
			continue
		}

		wordStatistic, ok := counter[lowerCaseWord]

		if ok {
			wordStatistic.Count++
		} else {
			counter[lowerCaseWord] = &WordStatistic{Word: lowerCaseWord, Count: 1}
		}
	}

	mapsValuesIterator := maps.Values(counter)

	return slices.Collect(mapsValuesIterator)
}

func firstWords(wordsStatistic []*WordStatistic, amount int) []string {
	topWords := make([]string, 0, amount)

	for count := 0; count < amount && count < len(wordsStatistic); count++ {
		word := wordsStatistic[count].Word
		topWords = append(topWords, word)
	}

	return topWords
}

func wordsStatisticComparer(a, b *WordStatistic) bool {
	if a.Count == b.Count {
		return a.Word < b.Word
	}

	return a.Count > b.Count
}
