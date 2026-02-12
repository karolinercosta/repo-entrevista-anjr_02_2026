package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

const DateFormat = "2006-01-02"

type Date struct {
	time.Time
}

func NewDate(year int, month time.Month, day int) Date {
	return Date{time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

// ParseDateOnly para o formato YYYY-MM-DD
func ParseDateOnly(s string) (Date, error) {
	t, err := time.Parse(DateFormat, s)
	if err != nil {
		return Date{}, err
	}
	return Date{t}, nil
}

// MarshalJSON implements json.Marshaler
func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.Format(DateFormat))
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	if s == "" {
		*d = Date{}
		return nil
	}
	parsed, err := time.Parse(DateFormat, s)
	if err != nil {
		return err
	}
	*d = Date{parsed}
	return nil
}

// String retorna a data (YYYY-MM-DD)
func (d Date) String() string {
	if d.IsZero() {
		return ""
	}
	return d.Format(DateFormat)
}

// MarshalBSONValue implements bson.ValueMarshaler for MongoDB
func (d Date) MarshalBSONValue() (bsontype.Type, []byte, error) {
	if d.IsZero() {
		return bson.MarshalValue(nil)
	}
	return bson.MarshalValue(d.Time)
}

func (d *Date) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	var tm time.Time
	err := bson.UnmarshalValue(t, data, &tm)
	if err != nil {
		return err
	}
	*d = Date{tm}
	return nil
}
