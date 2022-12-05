// Package main implements a simple server to return color dictates to esp8266
// controllers of neo-Pixel (ws2812 type) LED entities.
package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	log "github.com/golang/glog"
)

const (
	quiet = time.Duration
)

var (
	port = flag.Int("port", 6789, "Port which the server listens.")
	host = flag.String("host", "127.0.0.1", "Host/ip to listen upon.")

	// colorDicates is a simple slice of colors or patterns which the
	// fastLED library can encode to an LED entity.
	colorDictates = []string{
		"red",
		"orange",
		"yellow",
		"green",
		"blue",
		"indigo",
		"violet",
		"rainbow",
	}
)

const (
	// statusTmpl is the: timestamp(nanos), dictate to which LED entities should change.
	statusTmpl = "%d, %s\n"
)

// handler is the base struct used to handle http services.
type handler struct {
	dictate   string
	timestamp time.Time
	port      int
}

func newHandler(port int) (*handler, error) {
	rand.Seed(time.Now().UnixNano())
	return &handler{
		dictate:   pickDictate(),
		timestamp: time.Now(),
		port:      port,
	}, nil
}

// updateDicate changes the dictate if there's not been a change in duration (3 seconds).
func (h *handler) updateDictate(d time.Duration) {
	for {
		if h.timestamp.Since(time.Now()) > d {
			h.dictate = h.pickDictate()
		}
	}
}

func (h *handler) pickDictate() string {
	return colorDictates[rand.Intn(len(colorDictates))]
}

// status returns the current timestamped color dictate to client LED entities.
func (h *handler) status(w http.ResponseWriter, r *http.Request) {
	log.Info("Got status request")
	color := h.pickDictate()
	fmt.Fprintf(w, statusTmpl, time.Now().UnixNano(), color)
}

// update handles setting the current value for timestamp and color dictate.
func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	log.Info("Got update request")
	fmt.Fprintf(w, "Update message: %v\n", time.Now())
}

// index displays the selections to callers.
func (h *handler) index(w http.ResponseWriter, r *http.Request) {
	log.Info("Got index request")
	fmt.Fprintf(w, "Helo World: %v\n", time.Now())
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Infof("Got request for: %v", r.URL.Path)
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

	// Start a goroutine that will force a change

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", *host, *port),
		Handler: h,
	}
	log.Fatal(s.ListenAndServe())
}
