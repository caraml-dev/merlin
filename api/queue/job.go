package queue

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type Arguments map[string]interface{}

type Job struct {
	ID        int64     `json:"id" gorm:"primary_key;"`
	Arguments Arguments `json:"arguments"`
	Completed bool      `json:"completed"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (a Arguments) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Arguments) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}
