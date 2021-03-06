package util

func MinimumKeyValueOfMap(m *map[int64]bool) int64 {

	hasMin := false
	var min int64

	for k := range *m {
		if (*m)[k] && (!hasMin || k < min) {
			min = k
			hasMin = true
		}
	}

	return min
}
