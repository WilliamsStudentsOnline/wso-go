package user_home_info

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

// Expected TSV header from registrar WSO info dumps (e.g. Classes_of_YY_YY_WSO_Info.txt).
var expectedHeader = []string{
	"EMAIL_ADDR",
	"SUBOX",
	"DORM",
	"ENTRY",
	"CITY",
	"STATE",
	"COUNTRY",
	"POSTAL",
}

// Record is one row of registrar WSO student info.
type Record struct {
	UnixID      string
	Email       string
	SUBox       string
	Dorm        string
	Entry       string
	HomeTown    string
	HomeState   string
	HomeCountry string
	HomeZip     string
}

// ParseTSV reads a tab-separated registrar dump with the expected header.
func ParseTSV(r io.Reader) ([]Record, error) {
	reader := csv.NewReader(r)
	reader.Comma = '\t'
	reader.FieldsPerRecord = len(expectedHeader)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = false

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	if err := validateHeader(header); err != nil {
		return nil, err
	}

	var records []Record
	line := 1 // header consumed
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		line++
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}

		rec, err := parseRow(row)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		records = append(records, rec)
	}

	return records, nil
}

func validateHeader(header []string) error {
	if len(header) != len(expectedHeader) {
		return fmt.Errorf("header: expected %d columns, got %d", len(expectedHeader), len(header))
	}
	for i, want := range expectedHeader {
		got := strings.TrimSpace(header[i])
		// Tolerate UTF-8 BOM on the first column.
		got = strings.TrimPrefix(got, "\ufeff")
		if !strings.EqualFold(got, want) {
			return fmt.Errorf("header column %d: expected %q, got %q", i, want, got)
		}
	}
	return nil
}

func parseRow(row []string) (Record, error) {
	email := strings.TrimSpace(row[0])
	unixID, err := unixFromEmail(email)
	if err != nil {
		return Record{}, err
	}

	return Record{
		UnixID:      unixID,
		Email:       strings.ToLower(email),
		SUBox:       strings.TrimSpace(row[1]),
		Dorm:        strings.TrimSpace(row[2]),
		Entry:       strings.TrimSpace(row[3]),
		HomeTown:    strings.TrimSpace(row[4]),
		HomeState:   strings.TrimSpace(row[5]),
		HomeCountry: strings.TrimSpace(row[6]),
		HomeZip:     strings.TrimSpace(row[7]),
	}, nil
}

func unixFromEmail(email string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	const suffix = "@williams.edu"
	if !strings.HasSuffix(email, suffix) {
		return "", fmt.Errorf("email %q is not a @williams.edu address", email)
	}
	unix := strings.TrimSuffix(email, suffix)
	if unix == "" {
		return "", fmt.Errorf("email %q has empty local part", email)
	}
	return unix, nil
}

// NullableString returns nil for blank strings so GORM can write SQL NULL.
func NullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
