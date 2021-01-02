package sqlite

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/fzerorubigd/elosort/pkg/db"

	"github.com/stretchr/testify/suite"
)

type SQLiteTestSuit struct {
	suite.Suite
	dbx db.Storage
}

func TestNewSQLiteStorage(t *testing.T) {
	suite.Run(t, &SQLiteTestSuit{})
}

func (s *SQLiteTestSuit) SetupTest() {
	fl, err := ioutil.TempFile(os.TempDir(), "db_*.sqlite3")
	s.Require().NoError(err)
	path := fl.Name()
	s.Require().NoError(fl.Close())
	s.dbx, err = NewSQLiteStorage(context.Background(), path)
	s.Require().NoError(err)
}

func (s *SQLiteTestSuit) TearDownSuite() {
	s.Require().NoError(s.dbx.Close())
}

func (s *SQLiteTestSuit) TestAddItem() {
	item := db.Item{
		UserID:      100,
		Name:        "Test",
		Category:    10,
		Description: "Desc",
		URL:         "URL",
		Image:       "IMG",
		Rank:        1000,
	}
	ctx := context.Background()
	ret, err := s.dbx.Create(ctx, &item)
	s.Require().NoError(err)
	s.Assert().Greater(ret.ID, int64(0))

	get, err := s.dbx.GetByID(ctx, item.ID)
	s.Require().NoError(err)

	s.Assert().Equal(item, *get)

	get, err = s.dbx.GetByName(ctx, 100, 10, "Test")
	s.Require().NoError(err)
	s.Assert().Equal(item, *get)

	s.Require().NoError(s.dbx.SetRank(ctx, item.ID, 20000))
	get, err = s.dbx.GetByID(ctx, item.ID)
	s.Require().NoError(err)
	item.Rank = 20000
	item.Compared++
	s.Assert().Equal(item, *get)

	ret, err = s.dbx.Create(ctx, &item)
	s.Assert().Error(err)
	s.Assert().Nil(ret)

	item2 := db.Item{
		UserID:      100,
		Name:        "Test2",
		Category:    10,
		Description: "Desc",
		URL:         "URL",
		Image:       "IMG",
		Rank:        1000,
	}
	ret, err = s.dbx.Create(ctx, &item2)
	s.Require().NoError(err)
	s.Assert().Greater(ret.ID, int64(0))

	ret, err = s.dbx.GetByID(ctx, item2.ID)
	s.Require().NoError(err)
	s.Assert().Equal(item2, *ret)

	all, err := s.dbx.Items(ctx, item.UserID, 10, 1, 10)
	s.Require().NoError(err)

	s.Assert().Equal([]*db.Item{&item, &item2}, all)

	randoms, err := s.dbx.Random(ctx, 100, 10, 1)
	s.Require().NoError(err)

	s.Require().Len(randoms, 1)
	s.Require().Contains([]*db.Item{&item2}, randoms[0])

	randoms, err = s.dbx.Random(ctx, 100, 10, 2)
	s.Require().NoError(err)
	s.Require().Len(randoms, 2)
	s.Require().Contains([]*db.Item{&item, &item2}, randoms[0])
	s.Require().Contains([]*db.Item{&item, &item2}, randoms[1])
	s.Require().NotEqual(randoms[1], randoms[0])
}


func (s *SQLiteTestSuit) TestCategory() {
	cat := &db.Category{
		UserID:      100,
		Name:        "Name",
		Description: "",
	}

	ctx := context.Background()
	_, err := s.dbx.CreateCategory(ctx, cat)
	s.Require().Error(err)

	cat2 := &db.Category{
		UserID:      100,
		Name:        "Name 2",
		Description: "",
	}
	_, err = s.dbx.CreateCategory(ctx, cat2)
	s.Require().Error(err)

	cats , err := s.dbx.Categories(ctx, 100)
	s.Require().Error(err)
	s.Assert().Equal([]*db.Category{cat, cat2}, cats)
}