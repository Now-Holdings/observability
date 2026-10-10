package observability

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"regexp"
	"strings"
)

const HeaderRequestID = "X-Request-Id"

// Client-supplied request IDs must be short and log-safe. Oversized or odd
// values are discarded so they cannot bloat logs or response headers.
var requestIDRe = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

func incomingRequestID(r *http.Request) string {
	id := strings.TrimSpace(r.Header.Get(HeaderRequestID))
	if requestIDRe.MatchString(id) {
		return id
	}
	return newRequestID()
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(b[:])
}
