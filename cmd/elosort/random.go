package main

import (
	"errors"
	"flag"
	"math/rand"
	"time"

	"elbix.dev/elosort/pkg/store"
)

func init() {
	addCommand("random", "Compare two random items", compareRandom)
}

func compareRandom(store store.Interface, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	list, err := store.Load()
	if err != nil {
		return err
	}

	if len(list.Items) < 2 {
		return errors.New("not enough item")
	}
	rand.Seed(time.Now().UnixNano())
	first := rand.Intn(len(list.Items))
	var last int
	for {
		last = rand.Intn(len(list.Items))
		if last != first {
			break
		}
	}

	if first > last {
		first , last = last, first
	}
	return compareIndex(store, list, first, last)
}
