package internal

import "strings"

func MatchesAny(value string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}

	for _, filter := range filters {
		if strings.EqualFold(value, filter) {
			return true
		}
	}
	return false
}

func HasOverlap(values []string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}

	for _, value := range values {
		for _, filter := range filters {
			if strings.EqualFold(value, filter) {
				return true
			}
		}
	}
	return false
}
