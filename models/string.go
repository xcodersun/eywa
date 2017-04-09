package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"strings"
)

type StringMap map[string]string

func (m *StringMap) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), m)
}

func (m StringMap) Value() (driver.Value, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

type StringSlice []string

func (s *StringSlice) Scan(src interface{}) error {
	asBytes, ok := src.([]byte)
	if !ok {
		return errors.New("scan source was not []bytes")
	}

	asString := string(asBytes)
	if len(asString) == 0 {
		(*s) = StringSlice([]string{})
	} else {
		parsed := strings.Split(asString, ",")
		(*s) = StringSlice(parsed)
	}

	return nil
}

func (s StringSlice) Value() (driver.Value, error) {
	return strings.Join(s, ","), nil
}