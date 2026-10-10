package observability

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

const HeaderRequestID = "X-Request-Id"

func incomingRequestID(r *http.Request) string {
	id := strings.TrimSpace(r.Header.Get(HeaderRequestID))
	if id == "" {
		id = strings.TrimSpace(r.Header.Get("X-Cloud-Trace-Context"))
		if i := strings.IndexByte(id, '/'); i > 0 {
			id = id[:i]
		}
	}
	if id == "" {
		id = newRequestID()
	}
	return id
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(b[:])
}
