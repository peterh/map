package main

import (
	"fmt"
	"os"
)

func usage() {
	fmt.Println("Usage: map <configfile>")
	os.Exit(1)
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}

	m, err := read(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	m.draw()
	err = m.write()
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}
