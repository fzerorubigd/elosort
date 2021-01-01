package sqlite

import (
	"context"

	migrate "github.com/rubenv/sql-migrate"

	"github.com/fzerorubigd/elosort/pkg/db"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

type storage struct {
	db *sqlx.DB
}

func (s *storage) Close() error {
	return s.db.Close()
}

func (s *storage) Create(ctx context.Context, item *db.Item) (*db.Item, error) {
	q := `INSERT INTO items (user_id, name, description, url, image, rank) 
VALUES (:user_id, :name, :description, :url, :image, :rank)`
	res, err := s.db.NamedExecContext(ctx, q, item)
	if err != nil {
		return nil, errors.Wrap(err, "insert failed")
	}

	item.ID, err = res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *storage) GetByID(ctx context.Context, id int64) (*db.Item, error) {
	var item db.Item

	if err := s.db.GetContext(ctx, &item, "SELECT * FROM items WHERE id = $1", id); err != nil {
		return nil, errors.Wrap(err, "select failed")
	}

	return &item, nil
}

func (s *storage) GetByName(ctx context.Context, userID int64, name string) (*db.Item, error) {
	var item db.Item

	if err := s.db.GetContext(
		ctx,
		&item,
		"SELECT * FROM items WHERE user_id = $1 AND name = $2",
		userID, name); err != nil {
		return nil, errors.Wrap(err, "select failed")
	}

	return &item, nil
}

func (s *storage) SetRank(ctx context.Context, id int64, rank int) error {
	if _, err := s.db.ExecContext(ctx,
		"UPDATE items SET rank = $1, compared = compared + 1 WHERE id = $2",
		rank, id); err != nil {
		return errors.Wrap(err, "update failed")
	}

	return nil
}

func (s *storage) Items(ctx context.Context, userID int64, page, count int) ([]*db.Item, error) {
	if page <= 0 {
		page = 1
	}
	if count < 0 || count > 100 {
		count = 10
	}
	start := (page - 1) * count
	q := "SELECT * FROM items WHERE user_id = $1 ORDER BY rank DESC LIMIT $2, $3"
	var list []*db.Item
	if err := s.db.SelectContext(ctx, &list, q, userID, start, start+count); err != nil {
		return nil, errors.Wrap(err, "list failed")
	}

	return list, nil
}

func (s *storage) Random(ctx context.Context, userID int64, count int) ([]*db.Item, error) {
	q := "SELECT * FROM items WHERE user_id = $1 ORDER BY compared, RANDOM() LIMIT $2"
	var items []*db.Item

	if err := s.db.SelectContext(ctx, &items, q, userID, count); err != nil {
		return nil, errors.Wrap(err, "select failed")
	}

	return items, nil
}

func (s *storage) Remove(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM items WHERE id = $1", id)
	return errors.Wrap(err, "failed to delete the item")
}

func (s *storage) initialize(ctx context.Context) error {
	_, err := migrate.Exec(s.db.DB, "sqlite3", migrations, migrate.Up)
	return errors.Wrap(err, "migration failed")
}

func NewSQLiteStorage(ctx context.Context, dbPath string) (db.Storage, error) {
	dbx, err := sqlx.Connect("sqlite3", dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open the db")
	}

	s := &storage{
		db: dbx,
	}

	if err := s.initialize(ctx); err != nil {
		return nil, err
	}

	return s, nil
}
