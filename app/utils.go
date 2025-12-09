package app

import "strings"

// filterStrings filters a slice of strings by a query (case-insensitive)
func filterStrings(items []string, query string) []string {
	if query == "" {
		return items
	}
	q := strings.ToLower(query)
	// Pre-allocate to avoid reallocations during append
	filtered := make([]string, 0, len(items))
	for _, item := range items {
		if strings.Contains(strings.ToLower(item), q) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

// moveCursor adjusts a cursor within bounds, returning the new position
func moveCursor(cursor, delta, max int) int {
	cursor += delta
	if cursor < 0 {
		return 0
	}
	if cursor >= max {
		return max - 1
	}
	return cursor
}
