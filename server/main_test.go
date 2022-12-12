package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIndex(t *testing.T) {
	tests := []struct {
		desc       string
		reqStr     string
		wantPrefix string
		wantCode   int
	}{{
		desc:       "Index /",
		reqStr:     "/",
		wantPrefix: "Helo World: ",
		wantCode:   http.StatusOK,
	}, {
		desc:       "Index /foobar",
		reqStr:     "/foobar",
		wantPrefix: "Helo World: ",
		wantCode:   http.StatusNotFound,
	}}

	for _, test := range tests {
		// Fake request for anything index-y.
		req, err := http.NewRequest("GET", test.reqStr, nil)
		if err != nil {
			t.Fatalf("[%v]: failed to setup request: %v", test.desc, err)
		}

		// ResponseRecorer, satisfy http.ResponseWriter to record the response.
		rr := httptest.NewRecorder()
		h, err := newHandler(9999)
		if err != nil {
			t.Fatalf("[%v]: failed to create handler: %v", test.desc, err)
		}
		// Call serve on the handler.
		h.ServeHTTP(rr, req)

		// Check status code is expected.
		if status := rr.Code; status != test.wantCode {
			t.Errorf("[%v]: handler returned wrong status code: got %v want %v",
				test.desc, status, test.wantCode)
		}

		// Validate that the start of the reply is as expected, only if the request was ok.
		if rr.Code == http.StatusOK {
			if !strings.HasPrefix(rr.Body.String(), test.wantPrefix) {
				t.Errorf("[%v]: got/want mismatch:\n\twant: %s\n\tgot: %s", test.desc, test.wantPrefix, rr.Body.String())
			}
		}
	}
}

// update basic like: /update/basic/<valid macaddr>
func TestUpdate(t *testing.T) {
	tests := []struct {
		desc       string
		reqStr     string
		wantPrefix string
		wantCode   int
	}{{
		desc:       "/update/basic known station",
		reqStr:     "/update/basic/8C:AA:B5:7A:BC:AD",
		wantPrefix: "success SetColor",
		wantCode:   http.StatusOK,
	}, {
		desc:     "/update too few slashes.",
		reqStr:   "/update",
		wantCode: http.StatusBadRequest,
	}, {
		desc:     "/update/basic too few slashes.",
		reqStr:   "/update/basic",
		wantCode: http.StatusBadRequest,
	}, {
		desc:     "/update/basic too many slashes.",
		reqStr:   "/update/basic/things/to/do/C8:AA:B5:7A:BC:DA",
		wantCode: http.StatusBadRequest,
	}, {
		desc:     "/update/basic unknown station",
		reqStr:   "/update/basic/C8:AA:B5:7A:BC:DA",
		wantCode: http.StatusBadRequest,
	}}

	for _, test := range tests {
		// Fake request for anything update/basic
		req, err := http.NewRequest("GET", test.reqStr, nil)
		if err != nil {
			t.Fatalf("[%v]: failed to setup request: %v", test.desc, err)
		}

		// Init client content, so to avoid panic.
		initClients(2)

		// ResponseRecorer, satisfy http.ResponseWriter to record the response.
		rr := httptest.NewRecorder()
		h, err := newHandler(9999)
		if err != nil {
			t.Fatalf("[%v]: failed to create handler: %v", test.desc, err)
		}
		// Call serve on the handler.
		h.ServeHTTP(rr, req)

		// Check status code is expected.
		if status := rr.Code; status != test.wantCode {
			t.Errorf("[%v]: handler returned wrong status code: got %v want %v",
				test.desc, status, test.wantCode)
		}

		// Validate that the start of the reply is as expected, only if the request was ok.
		if rr.Code == http.StatusOK {
			if !strings.HasPrefix(rr.Body.String(), test.wantPrefix) {
				t.Errorf("[%v]: got/want mismatch:\n\twant: %s\n\tgot: %s",
					test.desc, test.wantPrefix, rr.Body.String())
			}
		}
	}
}

// Status url: /status?id=8C:AA:B5:7A:BC:AD&leds=10&len=500
func TestStatus(t *testing.T) {
	tests := []struct {
		desc     string
		reqStr   string
		want     string
		wantCode int
	}{{
		desc:     "/status no id",
		reqStr:   "/status",
		wantCode: http.StatusBadRequest,
	}, {
		desc:     "/status invalid LED count",
		reqStr:   "/status?id=8C:AA:B5:7A:BC:AD&leds=ten",
		wantCode: http.StatusBadRequest,
	}, {
		desc:     "/status invalid stepLen",
		reqStr:   "/status?id=8C:AA:B5:7A:BC:AD&leds=10&len=five",
		wantCode: http.StatusBadRequest,
	}, {
		desc:     "/status unknown client",
		reqStr:   "/status?id=C8:AA:B5:7A:BC:DA&leds=10&len=500",
		wantCode: http.StatusBadRequest,
	}, {
		desc:     "/status known client expect good output.",
		reqStr:   "/status?id=8C:AA:B5:7A:BC:AD&leds=2&len=10",
		wantCode: http.StatusOK,
		want:     `{"TS":1670808464406994837,"Data":[{"Steps":1,"Colors":[16777215,16777215]}]}`,
	}}

	for _, test := range tests {
		// Fake request for anything update/basic
		req, err := http.NewRequest("GET", test.reqStr, nil)
		if err != nil {
			t.Fatalf("[%v]: failed to setup request: %v", test.desc, err)
		}

		// Init client content, so to avoid panic.
		initClients(2)

		// ResponseRecorer, satisfy http.ResponseWriter to record the response.
		rr := httptest.NewRecorder()
		h, err := newHandler(9999)
		if err != nil {
			t.Fatalf("[%v]: failed to create handler: %v", test.desc, err)
		}
		// Call serve on the handler.
		h.ServeHTTP(rr, req)

		// Check status code is expected.
		if status := rr.Code; status != test.wantCode {
			t.Errorf("[%v]: handler returned wrong status code: got %v want %v",
				test.desc, status, test.wantCode)
		}

		// Validate that the start of the reply is as expected, only if the request was ok.
		if rr.Code == http.StatusOK {
			if diff := cmp.Diff(rr.Body.String(), test.want); diff != "" {
				t.Errorf("[%v]: got/want mismatch got+/want-): %s",
					test.desc, diff)
			}
		}
	}
}
