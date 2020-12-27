package file

import (
	"encoding/json"
	"os"
	"path/filepath"

	"elbix.dev/elosort/pkg/models"
	"elbix.dev/elosort/pkg/store"
)

type fileLoader struct {
	fileName string
}

func (f *fileLoader) Load() (*models.List, error) {
	fl, err := os.Open(filepath.Clean(f.fileName))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = fl.Close()
	}()
	var list models.List
	if err := json.NewDecoder(fl).Decode(&list); err != nil {
		return nil, err
	}

	return &list, nil
}

func (f *fileLoader) Save(list *models.List) error {
	fl, err := os.Create(filepath.Clean(f.fileName))
	if err != nil {
		return err
	}
	defer func() {
		_ = fl.Close()
	}()
	return json.NewEncoder(fl).Encode(list)
}

// NewFileStore is for create a new file storage
func NewFileStore(path string) (store.Interface, error) {
	fl := &fileLoader{
		fileName: path,
	}
	_, err := os.Stat(filepath.Clean(path))
	if os.IsNotExist(err) {
		// Its ok, just create new one
		if err := fl.Save(&models.List{}); err != nil {
			return nil, err
		}
	}
	return fl, nil
}
