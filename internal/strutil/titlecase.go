package strutil

import "strings"

var stopWords = map[string]bool{
	"a": true, "an": true, "and": true, "as": true,
	"at": true, "but": true, "by": true, "for": true,
	"in": true, "nor": true, "of": true, "on": true,
	"or": true, "per": true, "the": true, "to": true, "vs": true, "with": true, "from": true,
}

func TitleCase(s string) string {
	words := strings.Fields(strings.ToLower(s))
	for i, w := range words {
		if i == 0 || i == len(words)-1 || !stopWords[w] {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
