package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/howeyc/gopass"
	"github.com/tisba/fritz-tls/fritzbox"
)

type configOptions struct {
	host              string
	adminPassword     string
	fullchain         string
	privatekey        string
	certificateBundle io.Reader
}

func main() {
	config := setupConfiguration()

	session, err := fritzbox.PerformLogin(config.host, config.adminPassword)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Login successful!")

	status, response, err := fritzbox.UploadCertificate(config.host, session, config.certificateBundle)
	if err != nil {
		log.Fatal(err)
	}

	if status {
		log.Println("TLS certificate installation successful!")
	} else {
		log.Println("TLS certificate installation not successful, check response")
		log.Println(response)
		os.Exit(1)
	}
}

func setupConfiguration() configOptions {
	var config configOptions

	flag.StringVar(&config.host, "host", "http://fritz.box", "FritzBox host")
	flag.StringVar(&config.adminPassword, "password", "", "Admin password")
	flag.StringVar(&config.fullchain, "fullchain", "", "path to full certificate chain")
	flag.StringVar(&config.privatekey, "key", "", "path to private key")
	flag.Parse()

	config.certificateBundle = io.MultiReader(readerFromFile(config.fullchain), readerFromFile(config.privatekey))

	if config.adminPassword == "" {
		config.adminPassword = getPasswdFromStdin()

		if config.adminPassword == "" {
			log.Fatal("Admin password requried!")
		}
	}

	return config
}

func getPasswdFromStdin() string {
	fmt.Printf("Password (will be masked): ")
	pass, err := gopass.GetPasswdMasked()

	if err != nil {
		log.Fatal(err)
	}

	return string(pass)
}

func readerFromFile(path string) io.Reader {
	reader, err := os.OpenFile(path, os.O_RDONLY, 0600)
	if err != nil {
		log.Fatal(err)
	}

	return reader
}
