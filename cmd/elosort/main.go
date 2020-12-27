package main

import (
	"flag"
	"log"
	"os"

	"elbix.dev/elosort/pkg/store/file"
)

func main() {
	var db string

	flag.StringVar(&db, "storage", os.Getenv("SORT_DB"), "the path to folder to load the data")
	flag.Parse()

	storage, err := file.NewFileStore(db)
	if err != nil {
		log.Fatal(err)
	}

	if err := dispatch(storage, flag.Args()...); err != nil {
		log.Fatal(err.Error())
	}
}
