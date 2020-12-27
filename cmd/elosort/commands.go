package main

import (
	"flag"
	"fmt"
	"os"

	"elbix.dev/elosort/pkg/store"
)

var (
	allCommand []command
)

type command struct {
	Name        string
	Description string

	Run func(store.Interface, ...string) error
}

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()

	fmt.Fprintf(flag.CommandLine.Output(), "Sub commands:\n")

	for i := range allCommand {
		fmt.Fprintf(flag.CommandLine.Output(), "  %s: %s\n", allCommand[i].Name, allCommand[i].Description)
	}

}

func dispatch(list store.Interface, args ...string) error {
	if len(args) < 1 {
		return fmt.Errorf("atleast one arg is required")
	}

	sub := args[0]
	for i := range allCommand {
		if sub == allCommand[i].Name {
			return allCommand[i].Run(list, args...)
		}
	}

	flag.Usage()
	return fmt.Errorf("invalid command")
}

func addCommand(name, description string, run func(store.Interface, ...string) error) {
	allCommand = append(allCommand, command{
		Name:        name,
		Description: description,
		Run:         run,
	})
}
