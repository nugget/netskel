package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

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

var (
	stdoutBuffer string
	DATAFILE     string
)

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

	assert.Equal(t, s.Hostname, clientGet(s.UUID, "hostname"))
	assert.Equal(t, s.Username, clientGet(s.UUID, "username"))
}

func TestSendRaw(t *testing.T) {
	clearStdout()
	s := newSession()

	s.SendRaw(DATAFILE)
	assert.Equal(t, "Hello, world!\n", stdoutBuffer, "File was not sent correctly")
}

func TestSendHexdump(t *testing.T) {
	clearStdout()
	s := newSession()

	s.SendHexdump(DATAFILE)
	assert.Equal(t, "48656c6c6f2c20776f726c64210a\n", stdoutBuffer, "File was not sent correctly")
}

func TestSendBase64(t *testing.T) {
	clearStdout()
	s := newSession()

	s.SendBase64(DATAFILE)
	assert.Equal(t, "SGVsbG8sIHdvcmxkIQo=\n", stdoutBuffer, "File was not sent correctly")
}

func createSampleFile(filename string) {
	data := []byte("Hello, world!\n")
	ioutil.WriteFile(filename, data, 0644)
}

func TestMain(m *testing.M) {
	CLIENTDB = "testing.db"
	DATAFILE = "sample.dat"

	createSampleFile(DATAFILE)

	code := m.Run()

	os.Remove(CLIENTDB)
	os.Remove(DATAFILE)

	os.Exit(code)
}
