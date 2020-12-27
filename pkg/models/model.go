package models

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"elbix.dev/elosort/pkg/ranking"
	"elbix.dev/elosort/pkg/ranking/elo"

	"github.com/go-playground/validator/v10"
)

// Item is a single item in the list to compare and sort
type Item struct {
	ID          string `json:"id" validate:"required,gte=3"`
	Name        string `json:"name" validate:"required,gte=3"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`

	Rank     int `json:"level"`
	Compared int `json:"compared"`
}

// Validate validate the structure
func (i *Item) Validate() error {
	return validator.New().Struct(i)
}

// List of items to manage
type List struct {
	Name  string `json:"name"`
	Items []Item `json:"items"`

	ranker ranking.Ranker
}

func (l *List) less(i, j int) bool {
	if l.Items[i].Rank != l.Items[j].Rank {
		return l.Items[i].Compared < l.Items[j].Compared
	}
	return l.Items[i].Rank < l.Items[j].Rank
}

func (l *List) ByName(name string) (int, error) {
	name = strings.ToLower(name)
	for i := range l.Items {
		if name == strings.ToLower(l.Items[i].Name) {
			return i, nil
		}
	}

	return -1, errors.New("not found")
}

// Balance the list by new score result
func (l *List) Balance(i, j int, scoreI, scoreJ float64) error {
	if l.ranker == nil {
		l.ranker = elo.NewEloRankerDefault()
	}

	ll := len(l.Items)
	if i >= ll || j >= ll {
		return errors.New("index out of range")
	}
	fmt.Println(l.Items[i].Name, "=>", scoreI)
	fmt.Println(l.Items[j].Name, "=>", scoreJ)
	rI, rJ, err := l.ranker.Calculate(l.Items[i].Rank, l.Items[j].Rank, scoreI, scoreJ)
	if err != nil {
		return err
	}

	l.Items[i].Rank, l.Items[j].Rank = rI, rJ

	sort.Slice(l.Items, l.less)
	return nil
}
