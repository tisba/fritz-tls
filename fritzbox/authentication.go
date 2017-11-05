package fritzbox

import (
	"crypto/md5" // nolint: gas
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// PerformLogin performs a login and returns SessionInfo including
// the session id (SID) on success
func PerformLogin(host string, adminPassword string) (SessionInfo, error) {
	session, err := fetchSessionInfo(host + "/login_sid.lua")
	if err != nil {
		log.Fatal(err)
	}

	response := buildReponse(session.Challenge, adminPassword)

	session, err = fetchSessionInfo(host + "/login_sid.lua?sid=" + session.SID + "&username=&response=" + response)
	if err != nil {
		return SessionInfo{}, err
	}
	if session.SID == "0000000000000000" {
		return SessionInfo{}, errors.New("Login not successful")
	}

	return session, nil
}

func fetchSessionInfo(url string) (SessionInfo, error) {
	resp, err := http.Get(url)
	if err != nil {
		return SessionInfo{}, err
	}

	defer resp.Body.Close() // nolint: errcheck

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return SessionInfo{}, err
	}

	var sessionInfo SessionInfo
	err = xml.Unmarshal(body, &sessionInfo)
	if err != nil {
		return SessionInfo{}, err
	}

	return sessionInfo, nil
}

func buildReponse(challenge string, password string) string {
	challengePassword := utf8ToUtf16(challenge + "-" + password)

	md5Response := md5.Sum([]byte(challengePassword)) // nolint: gas

	return challenge + "-" + fmt.Sprintf("%x", md5Response)
}
