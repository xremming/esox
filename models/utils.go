package models

func lastIndexRune(s string, r rune) int {
	index := -1
	for i, v := range s {
		if v == r {
			index = i
		}
	}

	return index
}

func cutRight(s string, sep rune) (string, string) {
	sepIndex := lastIndexRune(s, sep)
	if sepIndex < 0 {
		return s, ""
	}

	return s[:sepIndex], s[sepIndex+1:]
}
