package isbn

import (
	testify "github.com/stretchr/testify/assert"
	"testing"
)

func TestCleanIsbn(t *testing.T) {
	assert := testify.New(t)

	type caseStruct struct {
		isbn      string
		cleanIsbn string
	}

	cases := []caseStruct{
		{
			isbn:      "978-0131103627",
			cleanIsbn: "9780131103627",
		},
		{
			isbn:      "978-0-13-110362-7",
			cleanIsbn: "9780131103627",
		},
		{
			isbn:      "978 0 13 110362 7",
			cleanIsbn: "9780131103627",
		},
		{
			isbn:      "978-0 13 110362 7",
			cleanIsbn: "9780131103627",
		},
	}

	for _, testCase := range cases {
		assert.Equal(testCase.cleanIsbn, CleanISBN(testCase.isbn))
	}
}

func TestConvertIsbn10to13(t *testing.T) {
	assert := testify.New(t)

	type caseStruct struct {
		isbn10 string
		isbn13 string
	}

	cases := []caseStruct{
		{
			isbn10: "0131103628",
			isbn13: "9780131103627",
		},
		{
			isbn10: "0134190440",
			isbn13: "9780134190440",
		},
	}
	for _, testCase := range cases {
		assert.Equal(testCase.isbn13, ConvertIsbn10to13(testCase.isbn10))
	}
}
