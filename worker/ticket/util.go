package ticket

func truncateStringSlice(strings []string, limit int) ([]string, bool) {
	if limit <= 0 {
		return []string{}, true
	}

	total := 0
	for _, s := range strings {
		total += len(s)
	}

	if total <= limit {
		return strings, false
	}

	// Iterate backwards from the end to find the smallest trailing slice
	// whose cumulative length is at least 'limit'
	sum := 0
	for i := len(strings) - 1; i >= 0; i-- {
		sum += len(strings[i])
		if sum >= limit {
			return strings[i:], true
		}
	}

	// In theory, this should never be reached.
	return strings, true
}
