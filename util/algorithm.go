package util

func MinimumKeyValueOfMap(m *map[int64]bool) int64 {

	hasMin := false
	var min int64

	for k := range *m {
		if !hasMin || k < min {
			min = k
		}
	}

	return min
}
