package fritzbox

import (
	"io"
	"io/ioutil"
	"strings"
)

// UploadCertificate uploads certificate and privatekey, provided via data
// and installs it
func UploadCertificate(host string, session SessionInfo, data io.Reader) (bool, string, error) {
	extraParams := [][]string{
		{"sid", session.SID}, // it's important that sid is the first argument!
		{"BoxCertPassword", ""},
	}

	request, err := fileUploadRequest(host+"/cgi-bin/firmwarecfg", "POST", extraParams, "BoxCertImportFile", "boxcert.cer", "application/x-x509-ca-cert", data)
	if err != nil {
		return false, "", err
	}

	client := getHTTPClient()

	response, err := client.Do(request)
	if err != nil {
		return false, "", err
	}
	defer response.Body.Close() // nolint: errcheck

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, "", err
	}

	return checkTLSInstallSuccess(body), string(body), nil
}

func checkTLSInstallSuccess(response []byte) bool {
	res := string(response)
	for _, message := range tlsInstallationSuccessMessages {
		if strings.Contains(res, message) {
			return true
		}
	}
	return false
}
