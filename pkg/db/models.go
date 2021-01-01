package db

import (
	"github.com/go-playground/validator/v10"
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
