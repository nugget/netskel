package main

import (
	"fmt"
	"os"

	"github.com/boltdb/bolt"
)

var (
	BASEDIR string = "/usr/local/netskel"
)

func clientList() {
	db, err := bolt.Open(BASEDIR+"/clients.db", 0660, nil)
	if err != nil {
		fmt.Printf("Unable to open client database: %v\n", err)
		return
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			fmt.Printf("[%s]\n", name)

			b := tx.Bucket(name)
			c := b.Cursor()

			for k, v := c.First(); k != nil; k, v = c.Next() {
				fmt.Printf("   %s: %s\n", k, v)
			}

			return nil
		})
	})
}

func main() {
	command := os.Args[1]

	switch command {
	case "list":
		clientList()
	}

	os.Exit(0)
}
