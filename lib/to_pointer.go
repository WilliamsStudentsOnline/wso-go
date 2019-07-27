package lib

func IntToPtr(i int) *int {
	return &i
}

func StrToPtr(s string) *string {
	return &s
}

func BoolToPtr(b bool) *bool {
	return &b
}

// Either goes with string if not blank or with default otherwise
func StrDefaults(value, defaultVal string) string {
	if value == "" {
		return defaultVal
	}
	return value
}

// Either goes with string ptr if not nil or with default otherwise
func StrPtrDefaults(value, defaultVal *string) *string {
	if value == nil {
		return defaultVal
	}
	return value
}

// Either goes with int ptr if not nil or with default otherwise
func IntPtrDefaults(value, defaultVal *int) *int {
	if value == nil {
		return defaultVal
	}
	return value
}

// Either goes with bool ptr if not nil or with default otherwise
func BoolPtrDefaults(value, defaultVal *bool) *bool {
	if value == nil {
		return defaultVal
	}
	return value
}