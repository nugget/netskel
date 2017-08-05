package main

import (
	"os"
	"fmt"
	"strings"
)

func syntaxError() {
	fmt.Println("ERROR")
	os.Exit(1)
}

func netskelDB() {
	fmt.Println("Moo")
	os.Exit(0)
}

func fingerprint(method, filename string) {
	hash := "THISISAHASH"

	fmt.Println(hash)
}

func main() {
	if os.Args[0] != "server" {
		syntaxError()
	}

	nsCommand := strings.Split(os.Args[2], " ")
	command := nsCommand[0]

	//for index, arg := range nsCommand {
	//	fmt.Printf("%2d: %v\n", index, arg)
	//}

	switch command {
		case "netskeldb":
			netskelDB()

		case "sha1":
			filename := nsCommand[1]
			fingerprint("sha1", filename)

		default:
			syntaxError()
	}

	os.Exit(0)
}

