package main

import "github.com/jeyoungjung/zerocoin/cli"

func main() {
	defer db.Close()
	cli.Start()
}
