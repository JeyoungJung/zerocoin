package main

import (
	"fmt"

	"github.com/jeyoungjung/zerocoin/person"
)

func main() {
	jay := person.Person{}
	jay.SetDetails("jay", 19)
	fmt.Println("main jay", jay)
    fmt.Println(jay.Name())
}
