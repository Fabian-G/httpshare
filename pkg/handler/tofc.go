package handler

import (
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

var adrRegex = regexp.MustCompile(".*:[0-9]*")

// TrustOnFirstConnect is a middleware that only allows the first n clients to connect.
// Zero value is not usable.
type TrustOnFirstConnect struct {
	maxClients       int
	connectedClients map[string]struct{}
	mutex            sync.Mutex
}

// NewTrustOnFirstConnect creates a new TrustOnFirstConnect
func NewTrustOnFirstConnect(maxClients int) *TrustOnFirstConnect {
	return &TrustOnFirstConnect{
		maxClients:       maxClients,
		connectedClients: make(map[string]struct{}, 0),
		mutex:            sync.Mutex{},
	}
}

// Tofc wraps a given handler to be onlyaccessible by the first n clients.
func (t *TrustOnFirstConnect) Tofc(next http.HandlerFunc) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		t.mutex.Lock()
		accepted := false
		adr := r.RemoteAddr
		if !adrRegex.MatchString(adr) {
			log.Printf("%s addresse not matching format. Something is wrong here. Blocking request.", adr)
			http.Error(rw, "Unauthorized", http.StatusUnauthorized)
			accepted = false
		} else {
			adr = strings.Split(adr, ":")[0]
		}
		if _, ok := t.connectedClients[adr]; ok {
			accepted = true
		} else {
			if len(t.connectedClients) < t.maxClients {
				t.connectedClients[adr] = struct{}{}
				log.Printf("Trusting client %s. (%d/%d)", adr, len(t.connectedClients), t.maxClients)
				accepted = true
			} else {
				log.Printf("Blocking client %s. Maximum of %d reached", adr, t.maxClients)
				http.Error(rw, "Unauthorized", http.StatusUnauthorized)
				accepted = false
			}
		}
		t.mutex.Unlock()
		if accepted {
			next.ServeHTTP(rw, r)
		}
	}
}
