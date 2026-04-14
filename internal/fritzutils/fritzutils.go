package fritzutils

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"log"
	"net"
	"net/url"
	"os"
	"strings"
	"syscall"
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

func GetPasswdFromFile(path string) string {
	if path == "" {
		return ""
	}
	data, err := os.ReadFile(path)

	if err != nil {
		log.Fatal(err)
	}
	pass := strings.TrimSpace(string(data))

	return pass
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
		host = url.Host
		port = "443"
	}

	log.Printf("Checking certificate validity of domain %s via %s\n", domain, url)

	conn, err := tls.Dial("tcp", host+":"+port, &tls.Config{InsecureSkipVerify: true, ServerName: domain})
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			log.Fatal(err)
		} else {
			panic("Server doesn't support SSL certificate err: " + err.Error())
		}
	}
	defer conn.Close()

	// this is taken from an example at https://pkg.go.dev/crypto/tls#Config
	cs := conn.ConnectionState()
	opts := x509.VerifyOptions{
		DNSName:       cs.ServerName,
		Intermediates: x509.NewCertPool(),
	}
	for _, cert := range cs.PeerCertificates[1:] {
		opts.Intermediates.AddCert(cert)
	}
	_, err = cs.PeerCertificates[0].Verify(opts)
	if err != nil {
		return false, false, time.Time{}
	}

	expiry := conn.ConnectionState().PeerCertificates[0].NotAfter

	return expiry.After(time.Now().Add(minValidity)), true, expiry
}
