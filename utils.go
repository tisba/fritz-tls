package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/howeyc/gopass"
)

func getPasswdFromStdin() string {
	fmt.Printf("FRITZ!Box Admin Password (will be masked): ")
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
