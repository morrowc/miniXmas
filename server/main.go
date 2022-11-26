// Package main implements a simple server to return color dictates to esp8266
// controllers of neo-Pixel (ws2812 type) LED entities.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	port = flag.Int("port", 6789, "Port which the server listens.")
	host = flag.String("host", "127.0.0.1", "Host/ip to listen upon.")
)

// handler is the base struct used to handle http services.
type handler struct {
	port int
}

func newHandler(port int) (*handler, error) {
	return &handler{
		port: port,
	}, nil
}

// status returns the current timestamped color dictate to client LED entities.
func (h *handler) status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Status here: %v\n", time.Now())
}

// update handles setting the current value for timestamp and color dictate.
func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Update message: %v\n", time.Now())
}

// index displays the selections to callers.
func (h *handler) index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Helo World: %v\n", time.Now())
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/update"):
		h.update(w, r)
	case strings.HasPrefix(r.URL.Path, "/status"):
		h.status(w, r)
	default:
		h.index(w, r)
	}
}

func main() {
	flag.Parse()
	log.Infof("Server will listen on port: %d", *port)

	h, err := newHandler(*port)
	if err != nil {
		log.Fatalf("failed to create handler: %v", err)
	}

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", *host, *port),
		Handler: h,
	}
	log.Fatal(s.ListenAndServe())
}
