package main

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/blackjack/syslog"
	"golang.org/x/crypto/ssh"
)

var (
	AUTHKEYSFILE string = os.Getenv("HOME") + "/.ssh/authorized_keys"
	CLIENT       string = os.Getenv("SSH_CLIENT")
)

func Debug(format string, a ...interface{}) {
	syslog.Syslogf(syslog.LOG_NOTICE, format, a...)
}

func Log(format string, a ...interface{}) {
	syslog.Syslogf(syslog.LOG_NOTICE, format, a...)
}

func Warn(format string, a ...interface{}) {
	syslog.Syslogf(syslog.LOG_WARNING, format, a...)
}

func syntaxError() {
	fmt.Println("ERROR")
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

	fmt.Printf("%s\t%o\t*\t%d\t%x\n", trimmed, mode, file.Size(), hash)
}

func listDir(dirname string) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		Warn("Error reading directory %v", dirname)
		return
	}

	for _, file := range files {
		if file.Name() == ".git" {
			continue
		}

		fullname := dirname + "/" + file.Name()

		switch mode := file.Mode(); {
		case mode.IsDir():
			trimmed := strings.TrimPrefix(fullname, "./")
			fmt.Printf("%s/\t%d\t*\n", trimmed, 700)
			listDir(fullname)
		case mode.IsRegular():
			dbFileLine(fullname)
		}
	}
}

func netskelDB() {
	servername, _ := os.Hostname()
	now := time.Now().Format("Mon, 2 Jan 2006 15:04:05 UTC")

	os.Chdir("db")

	fmt.Printf("#\n# .netskeldb for %v\n#\n# Generated %v by %v\n#\n", CLIENT, now, servername)

	listDir(".")

	os.Exit(0)
}

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

func hexdump(filename string) {
	linelength := 30
	count := 0

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		Warn("Error trying to hexdump %v: %v", filename, err)
		os.Exit(1)
	}

	for _, c := range file {
		fmt.Printf("%02x", c)
		count += 1
		if count >= linelength {
			count = 0
			fmt.Printf("\n")
		}
	}
	fmt.Printf("\n")
}

func addKey(hostname string) {
	servername, err := os.Hostname()
	now := time.Now().Format("Mon Jan _2 15:04:05 2006")

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		Warn("Error generating private key: %v", err)
		os.Exit(1)
	}
	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	pemdata := pem.EncodeToMemory(privateKeyPEM)

	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		Warn("Error constructing public key: %v", err)
		os.Exit(1)
	}
	pubdata := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(pub)))

	Debug("Appending %d byte public key to %s for %s (%v)", len(pubdata), AUTHKEYSFILE, hostname, CLIENT)

	f, err := os.OpenFile(AUTHKEYSFILE, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		Warn("Error writing public key: %v", err)
		os.Exit(1)
	}

	defer f.Close()

	if _, err = f.WriteString("restrict " + pubdata + " " + hostname + " " + now + "\n"); err != nil {
		panic(err)
	}

	fmt.Printf("#\n# Netskel private key generated by %v for %v (%v)\n#\n", servername, hostname, CLIENT)
	fmt.Println(string(pemdata))
	//fmt.Println("-- ")
	//fmt.Println(string(pubdata))

	os.Exit(0)
}

func main() {
	syslog.Openlog("netskel-server", syslog.LOG_PID, syslog.LOG_USER)

	if os.Args[0] != "server" {
		syntaxError()
	}

	nsCommand := strings.Split(os.Args[2], " ")
	command := nsCommand[0]

	Log("netskel-server launched for %v with %v", CLIENT, nsCommand)

	//for index, arg := range nsCommand {
	//	fmt.Printf("%2d: %v\n", index, arg)
	//}

	switch command {
	case "netskeldb":
		netskelDB()

	case "sha1":
		filename := nsCommand[1]
		hash, _ := fingerprint(filename)
		fmt.Println(hash)

	case "sendfile":
		filename := nsCommand[1]
		hexdump(filename)

	case "addkey":
		key := nsCommand[1]
		addKey(key)

	default:
		syntaxError()
	}

	os.Exit(0)
}
