package lib

func StringsContains(slice []string, str string) bool {
	for _, val := range slice {
		if val == str {
			return true
		}
	}

	return false
}

func StringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
