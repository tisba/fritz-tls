package fritzutils

import (
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

func CheckCertValidity(url *url.URL, domain string, minValidity time.Duration) (bool, bool, time.Time) {
	host, port, err := net.SplitHostPort(url.Host)
	if err != nil {
		panic("URL cannot be parsed to host and port" + err.Error())
	}

	conn, err := tls.Dial("tcp", host+":"+port, nil)
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
