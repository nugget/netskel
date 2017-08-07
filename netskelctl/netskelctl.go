package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

//	"golang.org/x/crypto/ssh/terminal"

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

	var (
		uuidList  []string
		hostNames map[string]string
		lastTimes map[string]string
	)

	hostNames = make(map[string]string)
	hostWidth := 0

	lastTimes = make(map[string]string)
	lastWidth := 0

	err = db.View(func(tx *bolt.Tx) error {

		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			uuid := string(name)
			uuidList = append(uuidList, uuid)

			b := tx.Bucket(name)
			c := b.Cursor()

			for k, v := c.First(); k != nil; k, v = c.Next() {
				switch string(k) {
				case "hostname":
					hostNames[uuid] = string(v)
					if len(string(v)) > hostWidth {
						hostWidth = len(string(v))
					}
				case "lastSeen":
					v = transformKey(k, v)
					lastTimes[uuid] = string(v)
					if len(string(v)) > lastWidth {
						lastWidth = len(string(v))
					}
				}
			}

			return nil
		})
	})

	formatString := "%-36s  %-" + strconv.Itoa(hostWidth) + "s  %-" + strconv.Itoa(lastWidth) + "s\n"

	fmt.Printf(formatString, "Client ID", "Hostname", "Last Seen")
	fmt.Printf(formatString,
		strings.Repeat("=", 36),
		strings.Repeat("=", hostWidth),
		strings.Repeat("=", lastWidth),
	)

	for _, uuid := range uuidList {
		hostName := hostNames[uuid]
		lastSeen := lastTimes[uuid]

		fmt.Printf(formatString, uuid, hostName, lastSeen)
	}
}

func transformKey(k, v []byte) []byte {
	retbuf := v

	switch string(k) {
	case "created", "lastSeen":
		epoch, _ := strconv.ParseInt(string(v), 10, 64)
		retbuf = []byte(time.Unix(epoch, 0).Format("Mon Jan 2 2006 @ 15:04:05 MST"))
	}

	return retbuf
}

func clientInfo(search string) {
	db, err := bolt.Open(BASEDIR+"/clients.db", 0660, nil)
	if err != nil {
		fmt.Printf("Unable to open client database: %v\n", err)
		return
	}
	defer db.Close()

	widestK := 0

	err = db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			searchHit := false

			if strings.Contains(strings.ToLower(string(name)), strings.ToLower(search)) {
				searchHit = true
			}

			b := tx.Bucket(name)
			c := b.Cursor()

			for k, v := c.First(); k != nil; k, v = c.Next() {
				if len(string(k)) > widestK {
					widestK = len(string(k))
				}
				if strings.Contains(strings.ToLower(string(v)), strings.ToLower(search)) {
					searchHit = true
				}
			}

			if searchHit == true {
				fmt.Printf("[%s]\n", name)
				for k, v := c.First(); k != nil; k, v = c.Next() {
					v = transformKey(k, v)
					formatString := "  %-" + strconv.Itoa(widestK) + "s: %s\n"
					fmt.Printf(formatString, k, v)
				}
			}

			return nil
		})
	})
}

func main() {
	// width, height, _ := terminal.GetSize(0)

	command := os.Args[1]

	switch command {
	case "list":
		clientList()
	case "info":
		clientInfo(os.Args[2])
	}

	os.Exit(0)
}
