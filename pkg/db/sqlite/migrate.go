package sqlite

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	migrate "github.com/rubenv/sql-migrate"
)

type readSeeker struct {
	data *bytes.Buffer
}

func (r *readSeeker) Read(p []byte) (n int, err error) {
	return r.data.Read(p)
}

func (r *readSeeker) Seek(offset int64, whence int) (int64, error) {
	if offset != 0 || whence != 0 {
		return 0, errors.New("not supported")
	}

	return 0, nil
}

type inlineMigration []string

func (im inlineMigration) FindMigrations() ([]*migrate.Migration, error) {
	res := make([]*migrate.Migration, 0, len(im))
	for i := range im {
		reader := bytes.NewBufferString(im[i])
		m, err := migrate.ParseMigration(fmt.Sprintf("%d_migration", i), &readSeeker{
			data: reader,
		})
		if err != nil {
			return nil, err
		}

		res = append(res, m)
	}

	return res, nil
}
