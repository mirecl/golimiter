package linters

import (
	"strings"
	"unicode"
)

func GetSegments(ident string) (segments []string) {
	if ident == "" {
		return nil
	}

	runes := []rune(ident)
	i := 0

	for j, c := range runes {
		if j == 0 {
			continue
		}

		var newSeg bool
		prev := runes[j-1]

		if unicode.IsUpper(c) {
			newSeg =
				unicode.IsLower(prev) ||
					unicode.IsDigit(prev) ||
					(j+1 < len(runes) && unicode.IsLower(runes[j+1]))

		}

		if newSeg {
			segments = append(segments, string(runes[i:j]))
			i = j
		}
	}

	// append leftover
	return append(segments, string(runes[i:]))
}

func GetSegmentCount(text string) int {
	return len(GetSegments(text))
}

func FindIdentsWithPartialPrefix(prefixSource string, idents []string) (commonPrefix string, found []string) {
	if prefixSource == "" {
		panic("empty prefix")
	}

	prefixSegs := GetSegments(prefixSource)

	// initial run: find idents that start with first segment
	prefix := prefixSegs[0]
	for _, ident := range idents {
		if len(ident) >= len(prefix) && strings.EqualFold(prefix, ident[:len(prefix)]) {
			found = append(found, ident)
		}
	}

	if len(found) == 0 {
		return
	}
	commonPrefix = prefix

	// try to extend common prefix
outer:
	for i := 1; i < len(prefixSegs); i++ {
		// extend prefix
		extendedPrefix := strings.Join(prefixSegs[:i+1], "")

		// test every previously found ident
		for _, ident := range found {
			if len(ident) < len(extendedPrefix) || !strings.EqualFold(extendedPrefix, ident[:len(extendedPrefix)]) {
				break outer
			}
		}

		// extend common prefix if all previously elements have it
		commonPrefix = extendedPrefix
	}
	return commonPrefix, found
}
