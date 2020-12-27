package store

import (
	"elbix.dev/elosort/pkg/models"
)

// Interface is the storage interface
type Interface interface {
	Load() (*models.List, error)
	Save(list *models.List) error
}
