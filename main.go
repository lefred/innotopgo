package main

import (
	"fmt"
	"os"

	"github.com/lefred/innotopgo/db"
	"github.com/lefred/innotopgo/innotop"
	"github.com/lefred/innotopgo/parse"
)

func main() {
	var displaytype = "simple"
	if len(os.Args) < 2 {
		fmt.Println("Usage: innotopgo mysql://<username>:<password>@<host>:3306")
		os.Exit(1)
	}
	if len(os.Args) < 3 {
		displaytype = "normal"
	} else {
		displaytype = os.Args[2]
	}
	uri, err := parse.Parse(os.Args[1])
	if err != nil {
		innotop.ExitWithError(err)
	}
	mydb, err := db.Connect(uri)
	if err != nil {
		innotop.ExitWithError(err)
	}
	defer mydb.Close()
	err = innotop.Processlist(mydb, displaytype)
	if err != nil {
		innotop.ExitWithError(err)
	}
}
