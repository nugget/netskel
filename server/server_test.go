package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
