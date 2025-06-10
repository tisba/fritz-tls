package fritzbox

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"time"

	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

var tlsInstallationSuccessMessages = []string{
	"Das SSL-Zertifikat wurde erfolgreich importiert.", // DE
	"Import of the SSL certificate was successful.",    // EN
	"Il certificato SSL è stato importato.",            // IT
}

func (fb *FritzBox) getHTTPClient() *http.Client {
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	if fb.Insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // nolint: gas
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   2 * time.Minute,
	}

	return client
}

func fileUploadRequest(uri string, method string, params [][]string, paramName string, fileName string, mimeType string, data io.Reader) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, param := range params {
		_ = writer.WriteField(param[0], param[1])
	}

	part, err := createFormFile(writer, paramName, fileName, mimeType)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, data)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, err
}

func createFormFile(w *multipart.Writer, fieldname string, filename string, contenttype string) (io.Writer, error) {
	replacer := strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition",
		fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
			replacer.Replace(fieldname), replacer.Replace(filename)))
	h.Set("Content-Type", contenttype)

	return w.CreatePart(h)
}

// FritzBox use UTF16 LittleEndian (aka UCS-2LE)
func utf8ToUtf16(input string) string {
	e := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	t := e.NewEncoder()

	outstr, _, err := transform.String(t, input)
	if err != nil {
		log.Fatal(err)
	}

	return outstr
}
