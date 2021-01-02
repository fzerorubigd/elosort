package db

import (
	"context"
	"io"
)

// Storage is an interface to handle the store
type Storage interface {
	Create(ctx context.Context, item *Item) (*Item, error)
	CreateCategory(ctx context.Context, category *Category) (*Category, error)
	GetCategoryByName(ctx context.Context, userID int64, name string) (*Category, error)
	GetCategoryByID(ctx context.Context, id int64) (*Category, error)
	GetByID(ctx context.Context, id int64) (*Item, error)
	GetByName(ctx context.Context, userID, category int64, name string) (*Item, error)
	SetRank(ctx context.Context, id int64, rank int) error
	Items(ctx context.Context, userID, category int64, page, count int) ([]*Item, error)
	Categories(ctx context.Context, userID int64) ([]*Category, error)
	Random(ctx context.Context, userID, category int64, count int) ([]*Item, error)
	UserByID(ctx context.Context, id int64) (*User, error)
	CreateUser(ctx context.Context, usr *User) error
	UpdateConfig(ctx context.Context, id int64, config *UserConfig) error
	io.Closer
}
