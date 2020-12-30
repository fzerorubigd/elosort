package db

import (
	"github.com/go-playground/validator/v10"
)

// Item is a single item in the list to compare and sort
type Item struct {
	ID          int64  `json:"id" validate:"required" db:"id"`
	UserID      int64  `json:"user_id" validate:"required" db:"user_id"`
	Name        string `json:"name" validate:"required,gte=3" db:"name"`
	Category    string `json:"category" validate:"required,gte=3" db:"category"`
	Description string `json:"description,omitempty" db:"description"`
	URL         string `json:"url,omitempty" db:"url"`
	Image       string `json:"image,omitempty" db:"image"`
	Rank        int    `json:"level" db:"rank"`
	Compared    int    `json:"compared" db:"compared"`
}

// Validate validate the structure
func (i *Item) Validate() error {
	return validator.New().Struct(i)
}
