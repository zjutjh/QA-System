package model

import (
	"database/sql/driver"
	"strings"
)

type Question struct {
	ID      uint   `json:"qid",gorm:"primary_key"`
	Text    string `json:"text"`
	Options Array  `json:"options",gorm:"type:longtext"`
	Type    string `json:"type"`
	ListID  uint   `json:"-"`
}

type Array []string

func (m *Array) Scan(val interface{}) error {
	s := val.([]uint8)
	ss := strings.Split(string(s), "|")
	*m = ss
	return nil
}

func (m Array) Value() (driver.Value, error) {
	str := strings.Join(m, "|")
	return str, nil
}
