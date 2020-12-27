package main

import (
	"errors"
	"flag"
	"strings"

	"elbix.dev/elosort/pkg/models"
	"elbix.dev/elosort/pkg/store"

	"github.com/google/uuid"
)

func interactiveInput(item *models.Item) error {
	return nil
}

func addItem(store store.Interface, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	var (
		item        models.Item
		interactive bool
	)
	fs.StringVar(&item.ID, "id", "", "The item id, auto generated if empty")
	fs.StringVar(&item.Name, "name", "", "The item name, mandatory")
	fs.StringVar(&item.Description, "description", "", "The item description, optional")
	fs.StringVar(&item.URL, "url", "", "The item URL, optional")
	fs.BoolVar(&interactive, "interactive", false, "Interactive")

	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	list, err := store.Load()
	if err != nil {
		return err
	}

	item.ID = strings.Trim(item.ID, "\n\t ")
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	if interactive {
		if err := interactiveInput(&item); err != nil {
			return err
		}
	}

	if err := item.Validate(); err != nil {
		return err
	}

	item.Rank = 10000

	if _, err = list.ByName(item.Name); err == nil {
		return errors.New("duplicate item")
	}

	list.Items = append(list.Items, item)

	return store.Save(list)
}

func init() {
	addCommand("add", "Add an item to the list of ranking", addItem)
}
