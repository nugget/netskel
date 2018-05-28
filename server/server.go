package main

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/blackjack/syslog"
	"github.com/boltdb/bolt"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
)

// AUTHKEYSFILE stores the location of the ssh keys file to store newly-assigned client keys.
var AUTHKEYSFILE = os.Getenv("HOME") + "/.ssh/authorized_keys"

// CLIENTDB is the filename of the client database file.
var CLIENTDB = "clients.db"

var Send = fmt.Printf
var Sendln = fmt.Println

// Debug logs a debug message to syslog.
func Debug(format string, a ...interface{}) {
	syslog.Syslogf(syslog.LOG_DEBUG, format, a...)
}

// Log logs a normal message to syslog.
func Log(format string, a ...interface{}) {
	syslog.Syslogf(syslog.LOG_NOTICE, format, a...)
}

// Warn logs a warning to syslog.
func Warn(format string, a ...interface{}) {
	syslog.Syslogf(syslog.LOG_WARNING, format, a...)
}

// Fatal aborts all services.
func Fatal(format string, a ...interface{}) {
	syslog.Syslogf(syslog.LOG_CRIT, format, a...)
	os.Exit(1)
}

// Session holds details about a remote client when serving a reauest.
type session struct {
	RemoteAddr string
	UUID       string
	Username   string
	Hostname   string
	Command    string
}

func newSession() session {
	s := session{}
	s.RemoteAddr = strings.Split(os.Getenv("SSH_CLIENT"), " ")[0]

	return s
}

func (s *session) Parse(nsCommand []string) {
	s.UUID = "nouuid"
	s.Username = "user"
	s.Hostname = "unknown"
	s.Command = strings.ToLower(nsCommand[0])

	var (
		uuidPosition     int
		usernamePosition int
		hostnamePosition int
	)

	switch s.Command {
	case "addkey":
		usernamePosition = 1
		hostnamePosition = 2
	case "netskeldb":
		uuidPosition = 1
		usernamePosition = 2
		hostnamePosition = 3
	case "sendfile", "sendbase64":
		uuidPosition = 2
		usernamePosition = 3
		hostnamePosition = 4
	default:
		return
	}

	if uuidPosition > 0 && len(nsCommand) > uuidPosition {
		c, err := uuid.FromString(nsCommand[uuidPosition])
		if err != nil {
			Fatal("Unable to parse client-supplied UUID %v: %v", nsCommand[uuidPosition], err)
		} else {
			s.UUID = c.String()
		}
	}

	if usernamePosition > 0 && len(nsCommand) > usernamePosition {
		s.Username = nsCommand[usernamePosition]
	}

	if hostnamePosition > 0 && len(nsCommand) > hostnamePosition {
		s.Hostname = nsCommand[hostnamePosition]
	}
}

func (s *session) NetskelDB() {
	servername, _ := os.Hostname()
	now := time.Now().Format("Mon, 2 Jan 2006 15:04:05 UTC")

	Send("#\n# .netskeldb for %s at %v\n#\n# Generated %v by %v\n#\n", s.UUID, s.RemoteAddr, now, servername)

	// Force-inject the client itself
	dbDirLine("bin")
	dbFileLine("bin/netskel")

	os.Chdir("db")
	err := listDir(".")
	if err != nil {
		Warn("Error listing directory: %v", err)
	} else {
		Log("Sent netskeldb to %s@%s at %s (%s)", s.Username, s.Hostname, s.RemoteAddr, s.UUID)
	}
}

func (s *session) Heartbeat() {
	now := time.Now()
	secs := strconv.Itoa(int(now.Unix()))

	clientPut(s.UUID, "inet", s.RemoteAddr)
	clientPut(s.UUID, "lastSeen", secs)
	clientPut(s.UUID, "hostname", s.Hostname)
	clientPut(s.UUID, "username", s.Username)

	Debug("Stored heartbeat for %v", s.UUID)
}

func syntaxError() {
	Send("ERROR\n")
	os.Exit(1)
}

func dbFileLine(filename string) {
	file, err := os.Stat(filename)
	if err != nil {
		Warn("Error Stat %v: %v", filename, err)
		return
	}

	trimmed := strings.TrimPrefix(filename, "./")

	hash, _ := fingerprint(filename)

	mode := 0600
	if file.Mode()&0111 != 0 {
		mode = 0700
	}

	Send("%s\t%o\t*\t%d\t%x\n", trimmed, mode, file.Size(), hash)
}

func dbDirLine(directory string) {
	trimmed := strings.TrimPrefix(directory, "./")
	Send("%s/\t%d\t*\n", trimmed, 700)
}

func listDir(dirname string) error {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		Warn("Error reading directory %v", dirname)
		return err
	}

	for _, file := range files {
		if file.Name() == ".git" {
			continue
		}

		fullname := dirname + "/" + file.Name()

		switch mode := file.Mode(); {
		case mode.IsDir():
			dbDirLine(fullname)
			err := listDir(fullname)
			if err != nil {
				return err
			}
		case mode.IsRegular():
			dbFileLine(fullname)
		}
	}

	return nil
}

func (s *session) SendBase64(filename string) error {
	linelength := 76
	count := 0

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	str := base64.StdEncoding.EncodeToString(file)

	for _, c := range str {
		Send("%c", c)
		count++
		if count >= linelength {
			count = 0
			Send("\n")
		}
	}
	Send("\n")

	Log("Sent base64 %s (%d bytes) to %s@%s at %s (%s)", filename, len(file), s.Username, s.Hostname, s.RemoteAddr, s.UUID)

	return nil
}

func (s *session) SendHexdump(filename string) error {
	linelength := 30
	count := 0

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	for _, c := range file {
		Send("%02x", c)
		count++
		if count >= linelength {
			count = 0
			Send("\n")
		}
	}
	Send("\n")
	Log("Sent hexdump %s (%d bytes) to %s@%s at %s (%s)", filename, len(file), s.Username, s.Hostname, s.RemoteAddr, s.UUID)

	return nil
}

func (s *session) SendRaw(filename string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	Send("%v", string(file))
	Log("Sent raw %s (%d bytes) to %s@%s at %s (%s)", filename, len(file), s.Username, s.Hostname, s.RemoteAddr, s.UUID)

	return nil
}

func (s *session) AddKey() error {
	servername, err := os.Hostname()
	now := time.Now()
	nowFmt := now.Format("Mon Jan _2 15:04:05 2006")
	uuid, _ := uuid.NewV4()
	cuuid := uuid.String()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	pemdata := pem.EncodeToMemory(privateKeyPEM)

	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		return err
	}
	pubdata := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pub)))

	f, err := os.OpenFile(AUTHKEYSFILE, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.WriteString("restrict " + pubdata + " " + s.Hostname + " " + cuuid + " " + nowFmt + "\n"); err != nil {
		panic(err)
	}

	Send("#\n# Netskel private key generated by %v for %v (%v)\n#\n", servername, s.Hostname, s.RemoteAddr)
	Send("# CLIENT_UUID %s\n#\n", uuid)
	Sendln(string(pemdata))

	secs := strconv.Itoa(int(now.Unix()))

	clientPut(cuuid, "hostname", s.Hostname)
	clientPut(cuuid, "originalHostname", s.Hostname)
	clientPut(cuuid, "created", secs)

	Log("Added %d byte public key to %s for %s@%s (%v) uuid %s", len(pubdata), AUTHKEYSFILE, s.Username, s.Hostname, s.RemoteAddr, uuid)

	return nil
}

// fingerprint calculates the MD5 fingerprint of a file.
func fingerprint(filename string) ([]byte, error) {
	var result []byte

	file, err := os.Open(filename)
	if err != nil {
		return result, err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return result, err
	}

	return hash.Sum(result), nil
}

// clientPut stores a key/value in the client database.
func clientPut(uuid, key, value string) error {
	db, err := bolt.Open(CLIENTDB, 0660, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		Warn("clientPut %v error: %v", uuid, err)
		return fmt.Errorf("Can't open client db: %v", err)
	}
	defer db.Close()

	berr := db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(uuid))
		if err != nil {
			return fmt.Errorf("Can't create bucket for %s: %v", uuid, err)
		}

		perr := b.Put([]byte(key), []byte(value))
		return perr
	})

	if berr != nil {
		Warn("clientPut %v error: %v", uuid, err)
		return fmt.Errorf("clientPut %v error: %v", uuid, err)
	}

	return nil
}

// clientGet retrieves the value of a key in the client database.
func clientGet(uuid, key string) (retval string) {
	db, err := bolt.Open(CLIENTDB, 0660, &bolt.Options{})
	if err != nil {
		Warn("Unable to open client database: %v\n", err)
		return ""
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(uuid))
		if b == nil {
			return nil
		}
		v := b.Get([]byte(key))
		retval = string(v)
		return nil
	})

	return retval
}

func main() {
	syslog.Openlog("netskel-server", syslog.LOG_PID, syslog.LOG_USER)

	s := newSession()

	if os.Args[0] != "server" {
		syntaxError()
	}

	nsCommand := strings.Split(os.Args[2], " ")
	s.Command = strings.ToLower(nsCommand[0])

	Debug("Launched from %v with %v", s.RemoteAddr, nsCommand)

	switch s.Command {
	case "netskeldb":
		s.Parse(nsCommand)
		s.Heartbeat()
		s.NetskelDB()

	case "md5":
		filename := nsCommand[1]
		hash, err := fingerprint(filename)
		if err != nil {
			Fatal("Unable to determine fingerprint for %s: %v", filename, err)
		}
		Sendln(hash)

	case "sendfile":
		s.Parse(nsCommand)
		filename := nsCommand[1]

		if filename == "db/bin/netskel" {
			filename = "bin/netskel"
		}

		err := s.SendHexdump(filename)
		if err != nil {
			Warn("Unable to SendHexDump %s: %v", filename, err)
		}

	case "sendbase64":
		s.Parse(nsCommand)
		filename := nsCommand[1]

		if filename == "db/bin/netskel" {
			filename = "bin/netskel"
		}

		err := s.SendBase64(filename)
		if err != nil {
			Warn("Unable to SendBase64 %s: %v", filename, err)
		}

	case "rawclient":
		filename := "bin/netskel"
		err := s.SendRaw(filename)
		if err != nil {
			Warn("Unable to SendRaw %s: %v", filename, err)
		}

	case "addkey":
		s.Parse(nsCommand)
		err := s.AddKey()
		if err != nil {
			Warn("Error in AddKey: %v", err)
			os.Exit(1)
		}

	case "uname":
		s.Parse(nsCommand)
		uname := nsCommand[4]
		clientPut(s.UUID, "uname", uname)

	default:
		syntaxError()
	}

	os.Exit(0)
}
