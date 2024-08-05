package fritzbox

import "net/url"

// FritzBox holds general information about the FRITZ!Box to talk to
type FritzBox struct {
	Host            string
	User            string
	Insecure        bool
	Domain          string
	session         SessionInfo
	VerificationURL *url.URL
}

// SessionInfo holds information about
// the current authenticated fritzbox session
//
// We only need SID and Challenge currently.
type SessionInfo struct {
	SID       string `xml:"SID"`
	Challenge string `xml:"Challenge"`
}

func (s *SessionInfo) Valid() bool {
	return s.SID != "0000000000000000"
}
