package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/Fabian-G/httpshare/pkg/certs"
	"github.com/Fabian-G/httpshare/pkg/handler"
	"github.com/Fabian-G/httpshare/pkg/resolve"
)

var (
	inline    = flag.Bool("i", false, "If the served content should be marked as inline content (Displayed directly in browser instead of opening a download dialog).")
	limitFlag = flag.Int("l", -1, "Limit number of reqeusts to n")
	tofcFlag  = flag.Int("t", -1, "The first n clients are trusted. All other connections will be blocked. This is global and not per file.")
	port      = flag.Int("p", 8080, "The port the http server should listen on")
	encrypt   = flag.Bool("e", false, "Whether or not Transport encryption should be used. If set httpshare will generate a self signed certificate on startup.")
	resolveIP = flag.Bool("r", false, "If set, the generated URLs will contain your public IP Addresse. For that another server will be queried.")
	uploadDir = flag.String("d", "", "If set to a path, httpshare will enable receive Mode and an upload form will be presented at /upload. Downloads will be saved at specified path.")
	stdInName = flag.String("s", "stdin", "Filename to use for serving stdin.")
)

func assembleHandleFunc(endpointId string, tofcHandler *handler.TrustOnFirstConnect, endpoint http.HandlerFunc) http.HandlerFunc {
	handleFunc := endpoint
	if *limitFlag >= 0 {
		handleFunc = handler.LimitRequests(endpointId, *limitFlag, handleFunc)
	}
	if *tofcFlag >= 0 {
		handleFunc = tofcHandler.Tofc(handleFunc)
	}
	return handleFunc
}

func getProtocol() string {
	if *encrypt {
		return "https"
	}
	return "http"
}

func registerHandlers(myIP string) {
	tofcHandler := handler.NewTrustOnFirstConnect(*tofcFlag)
	if len(strings.TrimSpace(*uploadDir)) != 0 {
		enableReceiveMode(myIP, tofcHandler)
	}
	for _, file := range flag.Args() {
		registerFileServer(myIP, file, tofcHandler)
	}
}

func enableReceiveMode(myIP string, tofcHandler *handler.TrustOnFirstConnect) {
	http.HandleFunc("/upload", assembleHandleFunc("upload", tofcHandler, handler.ServeUploadPage(*uploadDir)))
	log.Printf("Upload interface available at %s://%s:%d%s\n", getProtocol(), myIP, *port, "/upload")
}

func registerFileServer(myIP string, file string, tofcHandler *handler.TrustOnFirstConnect) {
	if s, err := os.Stat(file); (os.IsNotExist(err) || !s.Mode().IsRegular()) && file != "-" {
		log.Fatalf("%s does not exist or is not a regular file", file)
	}
	id, err := rand.Int(rand.Reader, big.NewInt(0xFFFFFF))
	if err != nil {
		log.Fatalf("Failed to generate id for %s", file)
	}
	urlPath := fmt.Sprintf("/%06x", id)
	log.Printf("%s available at %s://%s:%d%s\n", file, getProtocol(), myIP, *port, urlPath)

	if file != "-" {
		http.HandleFunc(urlPath, assembleHandleFunc(file, tofcHandler, handler.ServeFile(file, *inline)))
	} else {
		http.HandleFunc(urlPath, assembleHandleFunc(file, tofcHandler, handler.ServeStdIn(*stdInName, *inline)))
	}
}

func createConfigDirIfNotExist() string {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		// No Config dir. Use tmp instead
		userConfigDir, err = ioutil.TempDir("", "httpshare_*")
		if err != nil {
			log.Fatal("Unable to create temporary directory for configuration.")
		}
	}
	httpShareConfigDir := path.Join(userConfigDir, "httpshare")
	if _, err := os.Stat(httpShareConfigDir); os.IsNotExist(err) {
		if err := os.Mkdir(httpShareConfigDir, 0777); err != nil {
			log.Fatal("Could not create config dir")
		}
	}
	return httpShareConfigDir
}

func main() {
	flag.Parse()
	resolver := resolve.IPResolver(&resolve.LocalIPResolver{})
	if *resolveIP {
		resolver = resolve.NewPublicIPResolver(resolver)
	}
	rawIP := resolver.Resolve()
	ipForURL := resolve.FormatIPForURL(rawIP)
	httpShareConfigDir := createConfigDirIfNotExist()

	if noActionSpecified() {
		log.Fatal("You need to specify at least one file. Or enable upload mode.")
	}
	registerHandlers(ipForURL)

	if *encrypt {
		cert, key, err := certs.GetCertificate(httpShareConfigDir, rawIP)
		if err != nil {
			log.Fatalf("Unable to create certificate: %s", err)
		}
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", *port), cert, key, http.DefaultServeMux))
	} else {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), http.DefaultServeMux))
	}
}

func noActionSpecified() bool {
	return flag.NArg() == 0 && len(*uploadDir) == 0
}
