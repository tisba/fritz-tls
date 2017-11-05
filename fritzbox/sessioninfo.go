package fritzbox

// SessionInfo holds information about
// the current authenticated fritzbox session
//
// We only need SID and Challenge currently.
type SessionInfo struct {
	SID       string `xml:"SID"`
	Challenge string `xml:"Challenge"`
}
