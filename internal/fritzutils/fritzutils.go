package fritzutils

import (
	"io"
	"log"
	"os"

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
