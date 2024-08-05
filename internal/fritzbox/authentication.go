package fritzbox

import (
	"crypto/md5" // nolint: gas
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const FRITZBOX_AUTHENTICATION_PATH = "/login_sid.lua?version=2"

// PerformLogin performs a login and returns SessionInfo including
// the session id (SID) on success
func (fb *FritzBox) PerformLogin(adminPassword string) error {
	session, err := fb.fetchSessionInfo()
	if err != nil {
		return err
	}

	response := buildResponse(session.Challenge, adminPassword)

	session, err = fb.sendAuthResponse(response)
	if err != nil {
		return err
	}
	if !session.Valid() {
		return errors.New("login not successful")
	}

	fb.session = session

	return nil
}

func (fb *FritzBox) CheckSession() (bool, error) {
	client := fb.getHTTPClient()

	requestBody := strings.NewReader("sid=" + fb.session.SID)

	resp, err := client.Post(fb.Host+FRITZBOX_AUTHENTICATION_PATH, "application/x-www-form-urlencoded", requestBody)
	if err != nil {
		return false, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close() // nolint: errcheck

	var sessionInfo SessionInfo
	err = xml.Unmarshal(body, &sessionInfo)
	if err != nil {
		return false, err
	}

	return sessionInfo.SID == fb.session.SID, nil
}

func (fb *FritzBox) fetchSessionInfo() (SessionInfo, error) {
	url := fb.Host + FRITZBOX_AUTHENTICATION_PATH

	resp, err := fb.getHTTPClient().Get(url)
	if err != nil {
		return SessionInfo{}, err
	}

	defer resp.Body.Close() // nolint: errcheck

	body, err := io.ReadAll(resp.Body)
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

func (fb *FritzBox) sendAuthResponse(challengeResponse string) (SessionInfo, error) {
	requestBody := strings.NewReader("username=" + fb.User + "&response=" + challengeResponse)

	resp, err := fb.getHTTPClient().Post(fb.Host+FRITZBOX_AUTHENTICATION_PATH, "application/x-www-form-urlencoded", requestBody)
	if err != nil {
		return SessionInfo{}, err
	}

	defer resp.Body.Close() // nolint: errcheck

	body, err := io.ReadAll(resp.Body)
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

func buildResponse(challenge string, password string) string {
	if strings.HasPrefix(challenge, "2$") {
		return buildResponsePbkdf2(challenge, password)
	} else {
		return buildResponseMd5(challenge, password)
	}
}

func buildResponseMd5(challenge string, password string) string {
	challengePassword := utf8ToUtf16(challenge + "-" + password)

	md5Response := md5.Sum([]byte(challengePassword)) // nolint: gas

	return challenge + "-" + fmt.Sprintf("%x", md5Response)
}

// challenge is in format: 2$<iter1>$<salt1>$<iter2>$<salt2>
func buildResponsePbkdf2(challenge string, password string) string {
	challengeParts := strings.Split(challenge, "$")

	iter1, err := strconv.Atoi(challengeParts[1])
	if err != nil {
		log.Fatalf("Failed to convert iter1 to int: %v", err)
	}
	salt1, err := hex.DecodeString(challengeParts[2])
	if err != nil {
		log.Fatalf("Failed to decode salt1 hex string: %v", err)
	}

	iter2, err := strconv.Atoi(challengeParts[3])
	if err != nil {
		log.Fatalf("Failed to convert iter2 to int: %v", err)
	}
	salt2, err := hex.DecodeString(challengeParts[4])
	if err != nil {
		log.Fatalf("Failed to decode salt2 hex string: %v", err)
	}

	hash1 := pbkdf2.Key([]byte(password), []byte(salt1), iter1, 32, sha256.New)
	hash2 := pbkdf2.Key(hash1, []byte(salt2), iter2, 32, sha256.New)

	return fmt.Sprintf("%x", salt2) + "$" + fmt.Sprintf("%x", hash2)
}
