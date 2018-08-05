package fritzbox

// FritzBox holds general information about the FRITZ!Box to talk to
type FritzBox struct {
	Host     string
	Insecure bool
	Domain   string
	TLSPort  int
	session  SessionInfo
}

// SessionInfo holds information about
// the current authenticated fritzbox session
//
// We only need SID and Challenge currently.
type SessionInfo struct {
	SID       string `xml:"SID"`
	Challenge string `xml:"Challenge"`
}
