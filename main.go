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
	"os/signal"

	"github.com/Fabian-G/httpshare/pkg/certs"
	"github.com/Fabian-G/httpshare/pkg/handler"
	"github.com/Fabian-G/httpshare/pkg/resolve"
)

var (
	inline    = flag.Bool("i", false, "If the served content should be marked as inline content (Displayed directly in browser instead of opening a download dialog).")
	limitFlag = flag.Int("l", -1, "Limit number of reqeusts to n")
	port      = flag.Int("p", 8080, "The port the http server should listen on")
	encrypt   = flag.Bool("e", false, "Whether or not Transport encryption should be used. If set httpshare will generate a self signed certificate on startup.")
	resolveIP = flag.Bool("r", false, "If set, the generated URLs will contain your public IP Addresse. For that another server will be queried.")
)

func assembleHandleFunc(file string) http.HandlerFunc {
	handleFunc := handler.ServeFile(file, *inline)
	if *limitFlag >= 0 {
		handleFunc = handler.LimitRequests(file, *limitFlag, handleFunc)
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
	for _, file := range flag.Args() {
		if s, err := os.Stat(file); os.IsNotExist(err) || !s.Mode().IsRegular() {
			log.Fatalf("%s does not exist or is not a regular file", file)
		}
		id, err := rand.Int(rand.Reader, big.NewInt(0xFFFFFF))
		if err != nil {
			log.Fatalf("Failed to generate id for %s", file)
			continue
		}
		urlPath := fmt.Sprintf("/%06x", id)
		log.Printf("%s available at %s://%s:%d%s\n", file, getProtocol(), myIP, *port, urlPath)
		http.HandleFunc(urlPath, assembleHandleFunc(file))
	}
}

func scheduleCleanupOnExit(tmpDir string) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		sig := <-c
		log.Printf("Received %s cleaning up.", sig)
		os.RemoveAll(tmpDir)
		os.Exit(0)
	}()
}

func main() {
	flag.Parse()
	resolver := resolve.IPResolver(&resolve.LocalIPResolver{})
	if *resolveIP {
		resolver = resolve.NewPublicIPResolver(resolver)
	}
	rawIP := resolver.Resolve()
	ipForURL := resolve.FormatIPForURL(rawIP)

	if flag.NArg() == 0 {
		log.Fatal("You need to specify at least one file")
	}
	registerHandlers(ipForURL)

	if *encrypt {
		tmpDir, err := ioutil.TempDir("", "httpshare_*")
		if err != nil {
			log.Fatal("Unable to create temporary directory")
		}
		scheduleCleanupOnExit(tmpDir)
		cert, key, err := certs.GenerateCert(tmpDir, rawIP)
		log.Fatal(http.ListenAndServeTLS(fmt.Sprintf(":%d", *port), cert, key, http.DefaultServeMux))
	} else {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), http.DefaultServeMux))
	}
}
