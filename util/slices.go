package util

func FindInStringSlice(slice []string, elem string) bool {

	for _, sliceElem := range slice {
		if elem == sliceElem {
			return true
		}
	}

	return false
}
