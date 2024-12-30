package fritzutils

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/howeyc/gopass"
)

func GetPasswdFromStdin() string {
	pass, err := gopass.GetPasswdMasked()

	if err != nil {
		log.Fatal(err)
	}

	return string(pass)
}

func ReaderFromFile(path string) io.Reader {
	reader, err := os.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}

	return reader
}

func OpenFileWithNewline(path string) io.Reader {
	file, err := os.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}

	var buffer bytes.Buffer
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		buffer.WriteString(scanner.Text())
		buffer.WriteByte('\n') // Preserve existing newlines
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	content := buffer.Bytes()

	if len(content) == 0 {
		log.Fatal("Emtpy content for '" + path + "'")
	}

	// ensure the file ends with a new line
	if content[len(content)-1] != '\n' {
		buffer.WriteByte('\n')
	}

	return &buffer
}

func CheckCertValidity(url *url.URL, domain string, minValidity time.Duration) (bool, bool, time.Time) {
	host, port, err := net.SplitHostPort(url.Host)
	if err != nil {
		panic("URL cannot be parsed to host and port" + err.Error())
	}

	conn, err := tls.Dial("tcp", host+":"+port, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		panic("Server doesn't support SSL certificate err: " + err.Error())
	}

	err = conn.VerifyHostname(domain)
	if err != nil {
		return false, false, time.Time{}
	}
	expiry := conn.ConnectionState().PeerCertificates[0].NotAfter

	return expiry.After(time.Now().Add(minValidity)), true, expiry
}
