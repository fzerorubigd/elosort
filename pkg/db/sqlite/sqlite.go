package sqlite

import (
	"context"

	"github.com/go-acme/lego/log"
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
	q := `INSERT INTO items (user_id, name, category, description, url, image, rank) 
VALUES (:user_id, :name, :category ,:description, :url, :image, :rank)`
	res, err := s.db.NamedExecContext(ctx, q, item)
	if err != nil {
		return nil, errors.Wrap(err, "insert item failed")
	}

	item.ID, err = res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return item, nil
}

func (s *storage) CreateCategory(ctx context.Context, category *db.Category) (*db.Category, error) {
	q := `INSERT INTO categories (user_id, name, description) 
VALUES (:user_id, :name ,:description)`
	res, err := s.db.NamedExecContext(ctx, q, category)
	if err != nil {
		return nil, errors.Wrap(err, "insert category failed")
	}

	category.ID, err = res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (s *storage) GetCategoryByName(ctx context.Context, userID int64, name string) (*db.Category, error) {
	var cat db.Category
	if err := s.db.GetContext(ctx,
		&cat,
		"SELECT * FROM categories WHERE user_id = $1 AND name = $2",
		userID, name); err != nil {
		return nil, err
	}

	return &cat, nil
}

func (s *storage) Categories(ctx context.Context, userID int64) ([]*db.Category, error) {
	q := "SELECT * FROM categories WHERE user_id = $1 ORDER BY id"
	var list []*db.Category
	if err := s.db.SelectContext(ctx, &list, q, userID); err != nil {
		return nil, errors.Wrap(err, "list categories failed")
	}

	return list, nil
}

func (s *storage) GetByID(ctx context.Context, id int64) (*db.Item, error) {
	var item db.Item

	if err := s.db.GetContext(ctx, &item, "SELECT * FROM items WHERE id = $1", id); err != nil {
		return nil, errors.Wrap(err, "select failed")
	}

	return &item, nil
}

func (s *storage) GetByName(ctx context.Context, userID, category int64, name string) (*db.Item, error) {
	var item db.Item

	if err := s.db.GetContext(
		ctx,
		&item,
		"SELECT * FROM items WHERE user_id = $1 AND name = $2 AND category = $3",
		userID, name, category); err != nil {
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

func (s *storage) Items(ctx context.Context, userID, category int64, page, count int) ([]*db.Item, error) {
	if page <= 0 {
		page = 1
	}
	if count < 0 || count > 100 {
		count = 10
	}
	start := (page - 1) * count
	q := "SELECT * FROM items WHERE user_id = $1 AND category = $2 ORDER BY rank DESC LIMIT $3, $4"
	var list []*db.Item
	if err := s.db.SelectContext(ctx, &list, q, userID, category, start, start+count); err != nil {
		return nil, errors.Wrap(err, "list failed")
	}

	return list, nil
}

func (s *storage) Random(ctx context.Context, userID, category int64, count int) ([]*db.Item, error) {
	q := "SELECT * FROM items WHERE user_id = $1 AND category = $2 ORDER BY compared, RANDOM() LIMIT $3"
	var items []*db.Item

	if err := s.db.SelectContext(ctx, &items, q, userID, category, count); err != nil {
		return nil, errors.Wrap(err, "select failed")
	}

	return items, nil
}

func (s *storage) Remove(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM items WHERE id = $1", id)
	return errors.Wrap(err, "failed to delete the item")
}

func (s *storage) UserByID(ctx context.Context, id int64) (*db.User, error) {
	q := "SELECT * FROM users WHERE id = $1"
	var usr db.User
	if err := s.db.GetContext(ctx, &usr, q, id); err != nil {
		return nil, err
	}

	return &usr, nil
}

func (s *storage) GetCategoryByID(ctx context.Context, id int64) (*db.Category, error) {
	q := "SELECT * FROM categories WHERE id = $1"
	var cat db.Category
	if err := s.db.GetContext(ctx, &cat, q, id); err != nil {
		return nil, err
	}

	return &cat, nil
}

func (s *storage) CreateUser(ctx context.Context, usr *db.User) error {
	_, err := s.db.NamedExecContext(ctx,
		"INSERT INTO users (id, config) VALUES (:id, :config)", usr)
	return errors.Wrap(err, "insert failed")
}

func (s *storage) UpdateConfig(ctx context.Context,  id int64, config *db.UserConfig) error {
	_, err := s.db.ExecContext(ctx, "UPDATE users SET config = $1 WHERE id = $2", config, id)
	return errors.Wrap(err, "update failed")
}

func (s *storage) initialize(_ context.Context) error {
	_, err := migrate.Exec(s.db.DB, "sqlite3", migrations, migrate.Up)
	return errors.Wrap(err, "migration failed")
}

func (s *storage) fixup(ctx context.Context) error {
	q := "SELECT DISTINCT user_id FROM items WHERE category = 0"
	var ids []int64

	if err := s.db.SelectContext(ctx, &ids, q); err != nil {
		return err
	}

	for _, id := range ids {
		cat, err := s.GetCategoryByName(ctx, id, "Wishlist")
		if err != nil {
			cat = &db.Category{
				UserID: id,
				Name:   "Wishlist",
			}
			if _, err := s.CreateCategory(ctx, cat); err != nil {
				return err
			}
		}

		if _, err := s.db.ExecContext(ctx,
			"UPDATE items SET category = $1 WHERE user_id = $2 AND category = 0",
			cat.ID,
			id); err != nil {
			return err
		}
	}
	return nil
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

	if err := s.fixup(ctx); err != nil {
		log.Print("Err on fixing old data", err)
	}

	return s, nil
}
