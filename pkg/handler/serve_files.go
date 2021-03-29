package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

// ServeFile serves stdin over http
func ServeStdIn(stdInName string, inline bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s requested stdIn %s\n", r.RemoteAddr, stdInName)
		w.Header().Set("Content-Disposition", fmt.Sprintf("%s; filename=%s", inlineString(inline), stdInName))
		_, err := io.Copy(w, os.Stdin)
		if err != nil {
			log.Printf("Could not write stdin to client fully. %s", err.Error())
			return
		}
	}
}
