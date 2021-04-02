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
		fmt.Println("No URI provided !")
		os.Exit(1)
	}
	if len(os.Args) < 3 {
		displaytype = "normal"
	} else {
		displaytype = os.Args[2]
	}
	uri := parse.Parse(os.Args[1])
	mydb := db.Connect(uri)
	defer mydb.Close()
	err := innotop.Processlist(mydb, displaytype)
	if err != nil {
		fmt.Printf("error during processlist: %s", err)
		os.Exit(1)
	}
}
