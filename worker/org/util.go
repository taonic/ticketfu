package org

import "sort"

func truncateStringMap(m map[int64]string, limit int) (map[int64]string, bool) {
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
