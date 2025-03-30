package util

import "sort"

func TruncateStringSlice(strings []string, limit int) ([]string, bool) {
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

func TruncateStringMap(m map[int64]string, limit int) (map[int64]string, bool) {
	if len(m) <= limit {
		return m, false
	}

	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	truncatedKeys := keys[len(keys)-limit:]

	newMap := make(map[int64]string, limit)
	for _, k := range truncatedKeys {
		newMap[k] = m[k]
	}

	return newMap, true
}
