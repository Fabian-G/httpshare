package handler

import (
	"fmt"
	"log"
	"net/http"
	"path"
)

// ServeFile serves the specified file over http
func ServeFile(file string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s requested file %s\n", r.RemoteAddr, file)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", path.Base(file)))
		http.ServeFile(w, r, file)
	}
}