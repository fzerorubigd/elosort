package db

import (
	"context"
	"io"
)

// Storage is an interface to handle the store
type Storage interface {
	Create(ctx context.Context, item *Item) (*Item, error)
	GetByID(ctx context.Context, id int64) (*Item, error)
	GetByName(ctx context.Context, userID int64, name string) (*Item, error)
	SetRank(ctx context.Context, id int64, rank int) error
	Items(ctx context.Context, userID int64, page, count int) ([]*Item, error)
	Random(ctx context.Context, userID int64, count int) ([]*Item, error)
	io.Closer
}
