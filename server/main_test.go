package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestIndex(t *testing.T) {
	tests := []struct {
		desc     string
		reqStr   string
		wantFile string
		wantCode int
	}{{
		desc:     "Index /",
		reqStr:   "/",
		wantFile: "./src/index.html",
		wantCode: http.StatusOK,
	}, {
		desc:     "JS /static/colorpicker.js",
		reqStr:   "/static/colorpicker.js",
		wantFile: "./src/static/colorpicker.js",
		wantCode: http.StatusOK,
	}, {
		desc:     "Index /foobar",
		reqStr:   "/foobar",
		wantCode: http.StatusNotFound,
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

		// Check if the response is as expected if we have a file to compare to.
		if test.wantFile != "" {
			want, err := ioutil.ReadFile(test.wantFile)
			if err != nil {
				t.Fatalf("[%v]: failed to read file: %v", test.desc, err)
			}
			if diff := cmp.Diff(rr.Body.String(), string(want)); diff != "" {
				t.Errorf("[%v]: got/want mismatch (-got +want):\n%s", test.desc, diff)
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
	}, {
		desc:     "/update/notknown unknown url",
		reqStr:   "/update/notknown",
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
		want:     `{"TS":1,"Data":[{"Steps":1,"Colors":[16777215,16777215]}]}`,
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

// UpdateHSV Uri: /update/hsvtime/8c:aa:b5:7a:bc:ad
func TestUpdateHSVTime(t *testing.T) {
	postData := `{"Steps":[{"time":1000,"color":{"$":{"h":29,"s":32,"v":100,"a":1},"initialValue":{"h":0,"s":0,"v":100,"a":1},"index":0}},{"time":1000,"color":{"$":{"h":126,"s":51,"v":100,"a":1},"initialValue":{"h":0,"s":0,"v":100,"a":1},"index":1}}]}`

	tests := []struct {
		desc     string
		reqStr   string
		postData string
		contentT string
		want     string
		wantCode int
	}{{
		desc:     "Good Request",
		reqStr:   "/update/hsvtime/8c:aa:b5:7a:bc:ad",
		postData: postData,
		contentT: "application/json",
		want:     "ok",
		wantCode: http.StatusOK,
	}, {
		desc:     "Bad Request- unknown client",
		reqStr:   "/update/hsvtime/8c:aa:b5:7a:cb:da",
		postData: postData,
		contentT: "application/json",
		wantCode: http.StatusBadRequest,
	}, {
		desc:     "Bad Request - wrong slashes",
		reqStr:   "/update/hsvtime",
		postData: postData,
		contentT: "application/json",
		wantCode: http.StatusBadRequest,
	}, {
		desc:     "Bad Request - wrong content-Type",
		reqStr:   "/update/hsvtime/8c:aa:b5:7a:bc:ad",
		postData: postData,
		contentT: "application/www-urlencoded-data",
		wantCode: http.StatusUnsupportedMediaType,
	}}

	for _, test := range tests {
		// Init client content, so to avoid panic.
		initClients(2)
		// Make the fake data to send into the test POST, as an io.Reader.
		reader := bytes.NewBufferString(test.postData)

		// Create a new request with that test data.
		req, err := http.NewRequest(http.MethodPost, test.reqStr, reader)
		if err != nil {
			t.Fatalf("[%v]: failed to setup request: %v", test.desc, err)
		}

		// Set the content-type header on the post so the receiver can unpackage it.
		req.Header.Set("Content-Type", test.contentT)

		// ResponseRecorder, satisfy http.ResponseWriter to record the response.
		rr := httptest.NewRecorder()

		// Create a new Handler.
		h, err := newHandler(9999)
		if err != nil {
			t.Fatalf("[%v]: failed to create handler: %v", test.desc, err)
		}
		// Call serve on the handler, with the test request.
		h.ServeHTTP(rr, req)
		result := rr.Result()
		// Read the body from the result/response.
		defer result.Body.Close()
		body, err := ioutil.ReadAll(result.Body)
		if err != nil {
			t.Fatalf("failed to read the result body: %v", err)
		}

		// Check status code is expected.
		if status := result.StatusCode; status != test.wantCode {
			t.Errorf("[%v]: handler returned wrong status code: got %v want %v, body: %s",
				test.desc, status, test.wantCode, string(body))
		}

		// Validate that the reply is as expected, only if the request was ok.
		if result.StatusCode == http.StatusOK {
			if diff := cmp.Diff(string(body), test.want); diff != "" {
				t.Errorf("[%v]: got/want mismatch got+/want-): %s",
					test.desc, diff)
			}
		}
	}
}

func TestUpdateRGBTime(t *testing.T) {
}
