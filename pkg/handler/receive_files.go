package handler

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
)

// Receive files over http form
func ServeUploadPage(destinationDir string) http.HandlerFunc {
	assertDestinationDirExists(destinationDir)
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handleUpload(destinationDir, w, r)
		} else if r.Method == "GET" {
			serveUploadPage(w)
		}
	}
}

func assertDestinationDirExists(dir string) {
	if s, err := os.Stat(dir); os.IsNotExist(err) || !s.Mode().IsDir() {
		log.Fatalf("Upload directory \"%s\" does not exist or is not a directory", dir)
	}
}

func serveUploadPage(w http.ResponseWriter) {
	page := `
	<!DOCTYPE html>
	<html>
		<head>
			<meta charset="utf-8" />
		</head>
		<body>
			<form action="upload" method="post" enctype="multipart/form-data">
				<label for="file">Filename:</label>
				<input type="file" name="file" id="file"><br>
				<input type="submit" name="submit" value="Submit">
			</form>
		</body>
	</html>
`
	_, err := w.Write([]byte(page))
	if err != nil {
		log.Print("Could not write response to client.")
	}
}

func handleUpload(dir string, w http.ResponseWriter, r *http.Request) {
	log.Printf("Do you want to accept file....")
	inputFile, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Could not parse form data", http.StatusBadRequest)
		return
	}

	fullPath := path.Join(dir, header.Filename)
	if consented := askAndWaitForUserConsent(fullPath, r); !consented {
		log.Printf("Upload of %s rejected", fullPath)
		http.Error(w, "Upload rejected.", http.StatusUnauthorized)
		return
	}

	outputFile, err := os.OpenFile(fullPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		log.Printf("Could not open file %s for writing: %s", fullPath, err.Error())
		http.Error(w, "There was an error during upload.", http.StatusInternalServerError)
		return
	}

	_, err = io.Copy(outputFile, inputFile)
	if err != nil {
		log.Printf("Could not write to file %s due to: %s", fullPath, err.Error())
		http.Error(w, "There was an error during upload.", http.StatusInternalServerError)
		return
	}

	log.Printf("Download of %s completed.", fullPath)
	sendSuccessMessage(w)
}

func sendSuccessMessage(w http.ResponseWriter) {
	page := `
	<!DOCTYPE html>
	<html>
		<head>
			<meta charset="utf-8" />
		</head>
		<body>
			<h1>Upload successful</h1>
		</body>
	</html>
	`
	_, err := w.Write([]byte(page))
	if err != nil {
		log.Print("Could not write response to client.")
	}
}

func askAndWaitForUserConsent(fullPath string, r *http.Request) bool {
	fmt.Printf("\n*******************\n")
	if _, err := os.Stat(fullPath); err == nil {
		fmt.Printf("Warning: File \"%s\" already exists in destination dir. If you continue the file will be overriden.\n", fullPath)
	}
	fmt.Printf("Upload reuested by %s. Do you want to accept file %s? (y/N) ", r.RemoteAddr, fullPath)
	reader := bufio.NewReader(os.Stdin)
	in, err := reader.ReadString('\n')
	in = in[:len(in)-1]
	if err != nil {
		log.Printf("There was an error reading consent from stdin. Assuming disapproval. %s", err.Error())
		return false
	}
	return in == "y" || in == "Y"
}
