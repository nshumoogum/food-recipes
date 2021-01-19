package helpers

import "strings"

// WordSeparator represents a string used to separate
// a list of words found in an array (of strings)
const WordSeparator = ","

// StringifyWords concatenates a list of strings (string array)
// into a single string with the WordSeparator defining where
// a word ends and new one begins
func StringifyWords(words []string) (w string) {
	for _, v := range words {
		w = w + v + WordSeparator
	}

	w = strings.TrimSuffix(w, WordSeparator)

	return w
}
