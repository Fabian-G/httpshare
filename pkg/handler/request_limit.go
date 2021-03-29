package handler

import (
	"log"
	"net/http"
	"sync"
)

// LimitRequests allows at most maxRequests to an endpoint
func LimitRequests(id string, maxRequests int, other http.HandlerFunc) http.HandlerFunc {
	requests := 0
	requestCountMutex := sync.Mutex{}

	return func(w http.ResponseWriter, r *http.Request) {
		requestCountMutex.Lock()
		if requests >= maxRequests {
			log.Printf("Request from %s on %s blocked. Limit reached.\n", r.RemoteAddr, id)
			http.Error(w, "Limit reached.", http.StatusUnauthorized)
			requestCountMutex.Unlock()
			return
		}
		requests++
		log.Printf("Request count for %s: %d of %d", id, requests, maxRequests)
		requestCountMutex.Unlock()
		other.ServeHTTP(w, r)
	}
}
