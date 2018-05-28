package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
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
	{
		"unrecognized command string will fail to parse",
		session{
			Command:  "unrecognized",
			UUID:     "nouuid",
			Username: "user",
			Hostname: "unknown"},
	},
	{
		"netskeldb this-is-a-malformed-uuid luser host.example.com",
		session{
			Command:  "netskeldb",
			UUID:     "nouuid",
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
			// RemoteAddr may have been legitimately populated if you are
			// running go test on a remote host via ssh.  We remove it here
			// so that we know what to compare for with the assertions.
			s.RemoteAddr = ""

			assert.Equal(t, tt.out, s, "The session structure was not Parse()d correctly.")
		})
	}
}

func TestNetskelDB(t *testing.T) {
	clearStdout()
	s := newSession()

	s.NetskelDB()

	assert.Contains(t, stdoutBuffer, "server.go", "Netskeldb was not generated correctly")
	assert.Contains(t, stdoutBuffer, "bin/", "Netskeldb was not generated correctly")
}

func TestListDirParent(t *testing.T) {
	// This will hit the ".git" special handling and directory handling code
	clearStdout()
	err := listDir("..")
	assert.Nil(t, err)
	assert.Contains(t, stdoutBuffer, "700", "I expected at least one directory")
	assert.NotContains(t, stdoutBuffer, ".git/", "The git directory should be ignored")
}

func TestListDirNotFound(t *testing.T) {
	clearStdout()
	err := listDir("/this/directory/does/not/exist")
	assert.True(t, os.IsNotExist(err))
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

	err := s.SendRaw(DATAFILE)
	assert.Nil(t, err, "SendRaw exited with an error")
	assert.Equal(t, "Hello, world!\n", stdoutBuffer, "File was not sent correctly")
}

func TestSendRawNotFound(t *testing.T) {
	clearStdout()
	s := newSession()

	err := s.SendRaw("/this/file/does/not/exist")
	assert.True(t, os.IsNotExist(err), "SendRaw somehow sent a non-existent file.")
}

func TestSendHexdump(t *testing.T) {
	clearStdout()
	s := newSession()

	err := s.SendHexdump(DATAFILE)
	assert.Nil(t, err, "SendHexdump exited with an error")
	assert.Equal(t, "48656c6c6f2c20776f726c64210a\n", stdoutBuffer, "File was not sent correctly")
}

func TestSendHexdumpNotFound(t *testing.T) {
	clearStdout()
	s := newSession()

	err := s.SendHexdump("/this/file/does/not/exist")
	assert.True(t, os.IsNotExist(err), "SendHexdump somehow sent a non-existent file.")
}

func TestSendBase64(t *testing.T) {
	clearStdout()
	s := newSession()

	err := s.SendBase64(DATAFILE)
	assert.Nil(t, err, "SendBase64 exited with an error")
	assert.Equal(t, "SGVsbG8sIHdvcmxkIQo=\n", stdoutBuffer, "File was not sent correctly.")
}

func TestSendBase64NotFound(t *testing.T) {
	clearStdout()
	s := newSession()

	err := s.SendBase64("/this/file/does/not/exist")
	assert.True(t, os.IsNotExist(err), "SendBase64 somehow sent a non-existent file.")
}

func TestAddKey(t *testing.T) {
	clearStdout()
	s := newSession()

	s.Hostname = "host.example.org"
	s.Username = "luser"

	err := s.AddKey()
	if err != nil {
		t.Log(err)
		t.Error("AddKey error")
	}

	keyfile, err := ioutil.ReadFile(AUTHKEYSFILE)
	if err != nil {
		t.Error("Unable to read AUTHKEYSFILE")
	}

	r := regexp.MustCompile(`restrict ssh-rsa ([^ ]+) ([^ ]+) ([^ ]+) `)
	matches := r.FindStringSubmatch(string(keyfile))
	s.UUID = matches[3]

	assert.Contains(t, stdoutBuffer, "Netskel private key generated", "key not transmitted properly")
	assert.Equal(t, s.Hostname, matches[2])
	assert.Equal(t, s.Hostname, clientGet(s.UUID, "hostname"))
	assert.Equal(t, s.Hostname, clientGet(s.UUID, "originalHostname"))
}

func TestMain(m *testing.M) {
	CLIENTDB = "testing.db"
	DATAFILE = "sample.dat"
	AUTHKEYSFILE = "testing_keys"

	ioutil.WriteFile(DATAFILE, []byte("Hello, world!\n"), 0644)
	ioutil.WriteFile(AUTHKEYSFILE, []byte{}, 0644)

	code := m.Run()

	os.Remove(CLIENTDB)
	os.Remove(DATAFILE)
	os.Remove(AUTHKEYSFILE)

	os.Exit(code)
}
