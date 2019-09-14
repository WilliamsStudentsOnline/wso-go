package lib

func StringsContains(slice []string, str string) bool {
	for _, val := range slice {
		if val == str {
			return true
		}
	}

	return false
}
