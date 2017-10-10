package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"

	"golang.org/x/crypto/ssh/terminal"
)

var (
	BASEDIR      string = "/usr/local/netskel"
	verbose      bool
	showDisabled bool
)

func Debug(format string, a ...interface{}) {
	if verbose {
		fmt.Printf(format+"\n", a...)
	}
}

func clientList() {
	screenWidth, _, _ := terminal.GetSize(0)

	if screenWidth < 80 {
		// Deal with it
	}

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

			b := tx.Bucket(name)
			c := b.Cursor()

			isDisabled := b.Get([]byte("disabled"))
			if isDisabled == nil || showDisabled {
				uuidList = append(uuidList, uuid)

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
			}

			return nil
		})
	})

	fmt.Println(screenWidth)

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
	case "created", "lastSeen", "disabled":
		epoch, _ := strconv.ParseInt(string(v), 10, 64)
		retbuf = []byte(time.Unix(epoch, 0).Format("Mon Jan 2 2006 @ 15:04:05 MST"))
	}

	return retbuf
}

func clientInfo(search string, days int) {
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

			if search == "" || strings.Contains(strings.ToLower(string(name)), strings.ToLower(search)) {
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

			if days > 0 {
				lastSeen, err := strconv.Atoi(string(b.Get([]byte("lastSeen"))))
				if err == nil {
					last64 := int64(lastSeen)
					timeDiff := time.Now().Unix() - last64
					daysDiff := timeDiff / 86400

					if daysDiff < int64(days) {
						Debug("%s seen %d days ago, skipping for audit", name, daysDiff)
						searchHit = false
					}
				}
			}

			isDisabled := b.Get([]byte("disabled"))
			if isDisabled != nil && !showDisabled {
				searchHit = false
			}

			if searchHit == true {
				fmt.Printf("[%s]\n", name)
				for k, v := c.First(); k != nil; k, v = c.Next() {
					v = transformKey(k, v)
					formatString := "  %-" + strconv.Itoa(widestK) + "s: %s\n"
					fmt.Printf(formatString, k, v)
				}
				fmt.Println("")
			}

			return nil
		})
	})
}

func clientPut(uuid, key, value string) {
	db, err := bolt.Open(BASEDIR+"/clients.db", 0660, nil)
	if err != nil {
		fmt.Printf("Unable to open client database: %v\n", err)
		return
	}
	defer db.Close()

	berr := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(uuid))
		if b == nil {
			return fmt.Errorf("Unknown host")
		}

		if value != "" {
			oldVal := string(b.Get([]byte(key)))
			perr := b.Put([]byte(key), []byte(value))
			Debug("%s: %v -> %v", key, oldVal, value)
			return perr
		} else {
			perr := b.Delete([]byte(key))
			return perr
		}
	})

	if berr != nil {
		fmt.Printf("%v\n", berr)
	}
}

func disableClient(uuid string) error {
	now := time.Now()
	secs := strconv.Itoa(int(now.Unix()))
	clientPut(uuid, "disabled", secs)
	return nil
}

func enableClient(uuid string) error {
	clientPut(uuid, "disabled", "")
	return nil
}

func deleteClient(uuid string) error {
	db, err := bolt.Open(BASEDIR+"/clients.db", 0660, nil)
	if err != nil {
		fmt.Printf("Unable to open client database: %v\n", err)
		return err
	}
	defer db.Close()

	berr := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte(uuid))
		return err
	})

	if berr != nil {
		fmt.Printf("%v\n", berr)
	}

	return berr
}

func getArg(pos int, def string) string {
	if len(flag.Args()) < pos {
		Debug("Can't get pos %d from len %d", pos, len(flag.Args()))
		return def
	}

	val := flag.Arg(pos)

	if val == "" {
		return def
	} else {
		return val
	}
}

func getArgInt(pos int, def int) int {
	strArg := getArg(pos, "")

	if strArg == "" {
		return def
	} else {
		i, err := strconv.Atoi(strArg)
		if err != nil {
			return def
		} else {
			return i
		}
	}
}

func Usage() {
	fmt.Println("usage: netskelctl [flags] <command>\n")
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Println("\nCommands:")
	fmt.Println("  list               List all known Netskel hosts")
	fmt.Println("  info <uuid|search> Show detailed info for single host")
	fmt.Println("  disable <uuid>     Disable single host")
	fmt.Println("  enable <uuid>      Disable single host")
	fmt.Println("  delete <uuid>      Delete single host")
	fmt.Println("  audit <days>       Show hosts not seen in <days> days")
	os.Exit(1)
}

func main() {
	var (
		help bool
	)
	flag.BoolVar(&help, "h", false, "Show this usage information")
	flag.BoolVar(&verbose, "v", false, "Show verbose output")
	flag.BoolVar(&showDisabled, "a", false, "Include all (disabled) hosts")
	flag.Parse()

	if len(flag.Args()) == 0 || help {
		Usage()
	}
	command := flag.Args()[0]

	switch command {
	case "list":
		clientList()
	case "info":
		clientInfo(getArg(1, "netskelnotfound"), 0)
	case "audit":
		days := getArgInt(1, 7)
		fmt.Printf("The following netskel hosts have not been seen in over %d days:\n\n", days)
		clientInfo("", days)
	case "disable":
		disableClient(getArg(1, "netskelnotfound"))
	case "enable":
		enableClient(getArg(1, "netskelnotfound"))
	case "delete":
		deleteClient(getArg(1, "netskelnotfound"))
	}

	os.Exit(0)
}
