package isbn

import "strings"

func CleanISBN(isbn string) string {
	return strings.ReplaceAll(strings.ReplaceAll(isbn, " ", ""), "-", "")
}

func ConvertIsbn10to13(isbn10 string) string {
	if len(isbn10) != 10 {
		return isbn10
	}
	// add 978 and drop last digit
	isbn10 = "978" + isbn10[:9]

	// sum all digits
	sum := 0
	for i, c := range isbn10 {
		digit := int(c - '0')
		if i%2 == 0 {
			sum += digit * 1
		} else {
			sum += digit * 3
		}
	}
	// find the check digit
	checkDigit := (10 - (sum % 10)) % 10
	return isbn10 + string(rune('0'+checkDigit))
}
