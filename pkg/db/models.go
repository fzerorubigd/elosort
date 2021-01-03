package db

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

// Item is a single item in the list to compare and sort
type Item struct {
	ID          int64  `json:"id" validate:"required" db:"id"`
	UserID      int64  `json:"user_id" validate:"required" db:"user_id"`
	Name        string `json:"name" validate:"required,gte=3" db:"name"`
	Category    int64  `json:"category" db:"category"`
	Description string `json:"description,omitempty" db:"description"`
	URL         string `json:"url,omitempty" db:"url"`
	Image       string `json:"image,omitempty" db:"image"`
	Rank        int    `json:"level" db:"rank"`
	Compared    int    `json:"compared" db:"compared"`
}

// Category of items
type Category struct {
	ID          int64  `json:"id" validate:"required" db:"id"`
	UserID      int64  `json:"user_id" validate:"required" db:"user_id"`
	Name        string `json:"name" validate:"required,gte=3" db:"name"`
	Description string `json:"description,omitempty" db:"description"`
}

// User is the single user in the system
type User struct {
	ID     int64       `json:"id" validate:"required" db:"id"`
	Config *UserConfig `json:"config" db:"config"`
}

// UserConfig is the user configuration to be stored in JSON format in db
type UserConfig struct {
	DefaultCatID int64  `json:"default_cat_id"`
	ShowTwoStep  bool   `json:"show_two_step"`
	Language     string `json:"language"`
}

// Scan is to read the data from database
func (uc *UserConfig) Scan(src interface{}) error {
	switch t := src.(type) {
	case string:
		return json.Unmarshal([]byte(t), uc)
	case []byte:
		return json.Unmarshal(t, uc)
	default:
		return errors.Errorf("invalid type: %T", src)
	}
}

// Value to return the data for database
func (uc *UserConfig) Value() (driver.Value, error) {
	return json.Marshal(uc)
}

// GetID returns the id of the category, 0 is default category
func (c *Category) GetID() int64 {
	if c == nil {
		return 0
	}

	return c.ID
}

// Validate validate the structure
func (i *Item) Validate() error {
	return validator.New().Struct(i)
}
