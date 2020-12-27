package handler

import (
	"fmt"
	"log"
	"net/http"
	"path"
)

func inlineString(inline bool) string {
	if inline {
		return "inline"
	}
	return "attachment"
}

// ServeFile serves the specified file over http
func ServeFile(file string, inline bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s requested file %s\n", r.RemoteAddr, file)
		w.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=%s", inlineString(inline), path.Base(file)))
		http.ServeFile(w, r, file)
	}
}