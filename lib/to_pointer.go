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