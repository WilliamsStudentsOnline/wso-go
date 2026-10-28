package lib

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
)

type CSV []string

func (c *CSV) Scan(value interface{}) error {
	var s string
	switch v := value.(type) {
	case []byte:
		s = string(v)
	case string:
		s = v
	default:
		return errors.New(fmt.Sprint("incorrect type", value))
	}

	*c = CSV(strings.Split(s, ","))
	return nil
}

func (c CSV) Value() (driver.Value, error) {
	// Empty array = nil value
	if len(c) == 0 {
		return nil, nil
	}

	return []byte(strings.Join(c, ",")), nil
}
