package lib

import "time"

func IntToPtr(i int) *int {
	return &i
}

func Int32ToPtr(i int32) *int32 {
	return &i
}

func StrToPtr(s string) *string {
	return &s
}

func BoolToPtr(b bool) *bool {
	return &b
}

func TruePtr() *bool {
	t := true
	return &t
}

func FalsePtr() *bool {
	f := false
	return &f
}

func Float64ToPtr(f float64) *float64 {
	return &f
}

func TimeToPtr(t time.Time) *time.Time {
	return &t
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

// Either goes with time ptr if not nil or with default otherwise
func TimePtrDefaults(value, defaultVal *time.Time) *time.Time {
	if value == nil {
		return defaultVal
	}
	return value
}
