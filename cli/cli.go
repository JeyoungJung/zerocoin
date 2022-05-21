package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/jeyoungjung/zerocoin/db"
	explorer "github.com/jeyoungjung/zerocoin/explorer/templates"
	"github.com/jeyoungjung/zerocoin/rest"
)

func usage() {
	fmt.Printf("Welcome to Zerocoin\n\n")
	fmt.Printf("Please use the following flags:\n\n")
	fmt.Printf("-port:		Set the PORT of the server\n")
	fmt.Printf("-mode:		Choose between 'html' and 'rest'\n\n")
	os.Exit(0)
}

func Start() {
	if len(os.Args) == 2 { // If the there is nothing after the, go run main.go, run usage
		usage()
	}
	db.InitDB()
	port := flag.Int("port", 4000, "Set port of the server")
	mode := flag.String("mode", "rest", "Choose between 'html' and 'rest'")

	flag.Parse()

	switch *mode {
	case "rest":
		rest.Start(*port)
	case "html":
		explorer.Start(*port)
	case "both":
		go explorer.Start(*port + 1000)
		rest.Start(*port)
	default:
		usage()
	}
}
