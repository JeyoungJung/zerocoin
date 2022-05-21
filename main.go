package main

import (
	"github.com/jeyoungjung/zerocoin/cli"
	"github.com/jeyoungjung/zerocoin/db"
)

func main() {
	defer db.CloseDatabase()

	cli.Start()
}
