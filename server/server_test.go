package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/stretchr/testify/assert"
)

func init() {
	Send = fakeSend
	Sendln = fakeSendln
}

func fakeSend(format string, a ...interface{}) (int, error) {
	stdoutBuffer += fmt.Sprintf(format, a...)
	return 0, nil
}

func fakeSendln(a ...interface{}) (int, error) {
	stdoutBuffer += fmt.Sprintln(a...)
	return 0, nil
}

var stdoutBuffer string

func clearStdout() {
	stdoutBuffer = ""
}

func TestHarness(t *testing.T) {
	assert.Equal(t, 1, 1, "Math stopped working.")
}

var parseTests = []struct {
	in  string
	out session
}{
	{
		"netskeldb 6ec558e1-5f06-4083-9070-206819b53916 luser host.example.com",
		session{
			Command:  "netskeldb",
			UUID:     "6ec558e1-5f06-4083-9070-206819b53916",
			Username: "luser",
			Hostname: "host.example.com"},
	},
	{
		"addkey luser host.example.com",
		session{
			Command:  "addkey",
			UUID:     "nouuid",
			Username: "luser",
			Hostname: "host.example.com"},
	},
	{
		"sendfile db/testfile 6ec558e1-5f06-4083-9070-206819b53916 luser host.example.com",
		session{
			Command:  "sendfile",
			UUID:     "6ec558e1-5f06-4083-9070-206819b53916",
			Username: "luser",
			Hostname: "host.example.com"},
	},
	{
		"sendbase64 db/testfile 6ec558e1-5f06-4083-9070-206819b53916 luser host.example.com",
		session{
			Command:  "sendbase64",
			UUID:     "6ec558e1-5f06-4083-9070-206819b53916",
			Username: "luser",
			Hostname: "host.example.com"},
	},
}

func TestParsing(t *testing.T) {

	for _, tt := range parseTests {
		t.Run(tt.in, func(t *testing.T) {
			s := newSession()
			nsCommand := strings.Split(tt.in, " ")
			s.Parse(nsCommand)

			assert.Equal(t, tt.out, s, "The session structure was not Parse()d correctly.")
		})
	}
}

func TestDB(t *testing.T) {
	clearStdout()

	s := newSession()
	s.NetskelDB()

	assert.Contains(t, stdoutBuffer, "server.go", "Netskeldb was not generated correctly")
	assert.Contains(t, stdoutBuffer, "bin/", "Netskeldb was not generated correctly")
}

func TestHeartbeat(t *testing.T) {
	clearStdout()

	s := newSession()
	s.UUID = "6ec558e1-5f06-4083-9070-206819b53916"
	s.Hostname = "host.example.org"
	s.Username = "luser"

	s.Heartbeat()

	assert.Equal(t, s.Hostname, getKeyValue(s.UUID, "hostname"))
	assert.Equal(t, s.Username, getKeyValue(s.UUID, "username"))
}

func getKeyValue(uuid, key string) (retval string) {
	db, err := bolt.Open(CLIENTDB, 0660, &bolt.Options{})
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	defer db.Close()

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(uuid))
		v := b.Get([]byte(key))
		retval = string(v)
		return nil
	})

	return retval
}

func TestMain(m *testing.M) {
	CLIENTDB = "testing.db"

	code := m.Run()

	os.Remove(CLIENTDB)

	os.Exit(code)
}
