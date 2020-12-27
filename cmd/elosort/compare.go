package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"elbix.dev/elosort/pkg/models"
	"elbix.dev/elosort/pkg/store"
)

func compareIndex(store store.Interface, list *models.List, i, j int) error {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("Give me a number between 0-10, zero means you like the %q 100%% more than %q, and 10 means the opposite: ", list.Items[j].Name, list.Items[i].Name)

	var score float64
	for {
		s, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		score, err = strconv.ParseFloat(strings.Trim(s, " \n\t"), 64)
		if err != nil {
			fmt.Printf("%q is not a number, try again:", s)
			continue
		}

		if score < 0 || score > 10 {
			fmt.Printf("%q is not a between 0-10, try again:", s)
			continue
		}

		break
	}

	left := 10 - score

	if err := list.Balance(i, j, score/10, left/10); err != nil {
		return err
	}

	return store.Save(list)
}

func compareItems(store store.Interface, args ...string) error {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	remaining := fs.Args()
	if len(remaining) != 2 {
		return errors.New("select two item name")
	}

	list, err := store.Load()
	if err != nil {
		return err
	}

	first, err := list.ByName(remaining[0])
	if err != nil {
		return fmt.Errorf("%q not found", remaining[0])
	}

	last, err := list.ByName(remaining[1])
	if err != nil {
		return fmt.Errorf("%q not found", remaining[1])
	}

	if first == last {
		return errors.New("both are the same item")
	}

	return compareIndex(store, list, first, last)
}

func init() {
	addCommand("compare", "Compare two items", compareItems)
}
