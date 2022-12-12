// Package main implements a simple server to return color dictates to esp8266
// controllers of neo-Pixel (ws2812 type) LED entities.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/golang/glog"
	"github.com/lucasb-eyer/go-colorful"
)

const (
	ledCount = 5

	// List of locations with to organize their clients
	GUTTER = 0
	TEST   = 1

	// Period of time between stagnant light coloration.
	idleTime = 120 * time.Second
)

var (
	port = flag.Int("port", 6789, "Port which the server listens.")
	host = flag.String("host", "127.0.0.1", "Host/ip to listen upon.")

	// colorDicates is a simple slice of colors or patterns which the
	// fastLED library can encode to an LED entity.
	colorDictates = map[string]RGBColor{
		"AliceBlue":                  0xF0F8FF, ///< @htmlcolorblock{F0F8FF}
		"Amethyst":                   0x9966CC, ///< @htmlcolorblock{9966CC}
		"AntiqueWhite":               0xFAEBD7, ///< @htmlcolorblock{FAEBD7}
		"Aqua":                       0x00FFFF, ///< @htmlcolorblock{00FFFF}
		"Aquamarine":                 0x7FFFD4, ///< @htmlcolorblock{7FFFD4}
		"Azure":                      0xF0FFFF, ///< @htmlcolorblock{F0FFFF}
		"Beige":                      0xF5F5DC, ///< @htmlcolorblock{F5F5DC}
		"Bisque":                     0xFFE4C4, ///< @htmlcolorblock{FFE4C4}
		"Black":                      0x000000, ///< @htmlcolorblock{000000}
		"BlanchedAlmond":             0xFFEBCD, ///< @htmlcolorblock{FFEBCD}
		"Blue":                       0x0000FF, ///< @htmlcolorblock{0000FF}
		"BlueViolet":                 0x8A2BE2, ///< @htmlcolorblock{8A2BE2}
		"Brown":                      0xA52A2A, ///< @htmlcolorblock{A52A2A}
		"BurlyWood":                  0xDEB887, ///< @htmlcolorblock{DEB887}
		"CadetBlue":                  0x5F9EA0, ///< @htmlcolorblock{5F9EA0}
		"Chartreuse":                 0x7FFF00, ///< @htmlcolorblock{7FFF00}
		"Chocolate":                  0xD2691E, ///< @htmlcolorblock{D2691E}
		"Coral":                      0xFF7F50, ///< @htmlcolorblock{FF7F50}
		"CornflowerBlue":             0x6495ED, ///< @htmlcolorblock{6495ED}
		"Cornsilk":                   0xFFF8DC, ///< @htmlcolorblock{FFF8DC}
		"Crimson":                    0xDC143C, ///< @htmlcolorblock{DC143C}
		"Cyan":                       0x00FFFF, ///< @htmlcolorblock{00FFFF}
		"DarkBlue":                   0x00008B, ///< @htmlcolorblock{00008B}
		"DarkCyan":                   0x008B8B, ///< @htmlcolorblock{008B8B}
		"DarkGoldenrod":              0xB8860B, ///< @htmlcolorblock{B8860B}
		"DarkGray":                   0xA9A9A9, ///< @htmlcolorblock{A9A9A9}
		"DarkGrey":                   0xA9A9A9, ///< @htmlcolorblock{A9A9A9}
		"DarkGreen":                  0x006400, ///< @htmlcolorblock{006400}
		"DarkKhaki":                  0xBDB76B, ///< @htmlcolorblock{BDB76B}
		"DarkMagenta":                0x8B008B, ///< @htmlcolorblock{8B008B}
		"DarkOcolorElementliveGreen": 0x556B2F, ///< @htmlcolorblock{556B2F}
		"DarkOrange":                 0xFF8C00, ///< @htmlcolorblock{FF8C00}
		"DarkOrchid":                 0x9932CC, ///< @htmlcolorblock{9932CC}
		"DarkRed":                    0x8B0000, ///< @htmlcolorblock{8B0000}
		"DarkSalmon":                 0xE9967A, ///< @htmlcolorblock{E9967A}
		"DarkSeaGreen":               0x8FBC8F, ///< @htmlcolorblock{8FBC8F}
		"DarkSlateBlue":              0x483D8B, ///< @htmlcolorblock{483D8B}
		"DarkSlateGray":              0x2F4F4F, ///< @htmlcolorblock{2F4F4F}
		"DarkSlateGrey":              0x2F4F4F, ///< @htmlcolorblock{2F4F4F}
		"DarkTurquoise":              0x00CED1, ///< @htmlcolorblock{00CED1}
		"DarkViolet":                 0x9400D3, ///< @htmlcolorblock{9400D3}
		"DeepPink":                   0xFF1493, ///< @htmlcolorblock{FF1493}
		"DeepSkyBlue":                0x00BFFF, ///< @htmlcolorblock{00BFFF}
		"DimGray":                    0x696969, ///< @htmlcolorblock{696969}
		"DimGrey":                    0x696969, ///< @htmlcolorblock{696969}
		"DodgerBlue":                 0x1E90FF, ///< @htmlcolorblock{1E90FF}
		"FireBrick":                  0xB22222, ///< @htmlcolorblock{B22222}
		"FloralWhite":                0xFFFAF0, ///< @htmlcolorblock{FFFAF0}
		"ForestGreen":                0x228B22, ///< @htmlcolorblock{228B22}
		"Fuchsia":                    0xFF00FF, ///< @htmlcolorblock{FF00FF}
		"Gainsboro":                  0xDCDCDC, ///< @htmlcolorblock{DCDCDC}
		"GhostWhite":                 0xF8F8FF, ///< @htmlcolorblock{F8F8FF}
		"Gold":                       0xFFD700, ///< @htmlcolorblock{FFD700}
		"Goldenrod":                  0xDAA520, ///< @htmlcolorblock{DAA520}
		"Gray":                       0x808080, ///< @htmlcolorblock{808080}
		"Grey":                       0x808080, ///< @htmlcolorblock{808080}
		"Green":                      0x008000, ///< @htmlcolorblock{008000}
		"GreenYellow":                0xADFF2F, ///< @htmlcolorblock{ADFF2F}
		"Honeydew":                   0xF0FFF0, ///< @htmlcolorblock{F0FFF0}
		"HotPink":                    0xFF69B4, ///< @htmlcolorblock{FF69B4}
		"IndianRed":                  0xCD5C5C, ///< @htmlcolorblock{CD5C5C}
		"Indigo":                     0x4B0082, ///< @htmlcolorblock{4B0082}
		"Ivory":                      0xFFFFF0, ///< @htmlcolorblock{FFFFF0}
		"Khaki":                      0xF0E68C, ///< @htmlcolorblock{F0E68C}
		"LavendcolorElementer":       0xE6E6FA, ///< @htmlcolorblock{E6E6FA}
		"LavenderBlush":              0xFFF0F5, ///< @htmlcolorblock{FFF0F5}
		"LawnGreen":                  0x7CFC00, ///< @htmlcolorblock{7CFC00}
		"LemonChiffon":               0xFFFACD, ///< @htmlcolorblock{FFFACD}
		"LightBlue":                  0xADD8E6, ///< @htmlcolorblock{ADD8E6}
		"LightCoral":                 0xF08080, ///< @htmlcolorblock{F08080}
		"LightCyan":                  0xE0FFFF, ///< @htmlcolorblock{E0FFFF}
		"LightGoldenrodYellow":       0xFAFAD2, ///< @htmlcolorblock{FAFAD2}
		"LightGreen":                 0x90EE90, ///< @htmlcolorblock{90EE90}
		"LightGrey":                  0xD3D3D3, ///< @htmlcolorblock{D3D3D3}
		"LightPink":                  0xFFB6C1, ///< @htmlcolorblock{FFB6C1}
		"LightSalmon":                0xFFA07A, ///< @htmlcolorblock{FFA07A}
		"LightSeaGreen":              0x20B2AA, ///< @htmlcolorblock{20B2AA}
		"LightSkyBlue":               0x87CEFA, ///< @htmlcolorblock{87CEFA}
		"LightSlateGray":             0x778899, ///< @htmlcolorblock{778899}
		"LightSlateGrey":             0x778899, ///< @htmlcolorblock{778899}
		"LightSteelBlue":             0xB0C4DE, ///< @htmlcolorblock{B0C4DE}
		"LightYellow":                0xFFFFE0, ///< @htmlcolorblock{FFFFE0}
		"Lime":                       0x00FF00, ///< @htmlcolorblock{00FF00}
		"LimeGreen":                  0x32CD32, ///< @htmlcolorblock{32CD32}
		"Linen":                      0xFAF0E6, ///< @htmlcolorblock{FAF0E6}
		"Magenta":                    0xFF00FF, ///< @htmlcolorblock{FF00FF}
		"Maroon":                     0x800000, ///< @htmlcolorblock{800000}
		"MediumAquamarine":           0x66CDAA, ///< @htmlcolorblock{66CDAA}
		"MediumBlue":                 0x0000CD, ///< @htmlcolorblock{0000CD}
		"MediumOrchid":               0xBA55D3, ///< @htmlcolorblock{BA55D3}
		"MediumPurple":               0x9370DB, ///< @htmlcolorblock{9370DB}
		"MediumSeaGreen":             0x3CB371, ///< @htmlcolorblock{3CB371}
		"MediumSlateBlue":            0x7B68EE, ///< @htmlcolorblock{7B68EE}
		"MediumSpringGreen":          0x00FA9A, ///< @htmlcolorblock{00FA9A}
		"MediumTurquoise":            0x48D1CC, ///< @htmlcolorblock{48D1CC}
		"MediumVioletRed":            0xC71585, ///< @htmlcolorblock{C71585}
		"MidnightBlue":               0x191970, ///< @htmlcolorblock{191970}
		"MintCream":                  0xF5FFFA, ///< @htmlcolorblock{F5FFFA}
		"MistyRose":                  0xFFE4E1, ///< @htmlcolorblock{FFE4E1}
		"Moccasin":                   0xFFE4B5, ///< @htmlcolorblock{FFE4B5}
		"NavajoWhite":                0xFFDEAD, ///< @htmlcolorblock{FFDEAD}
		"Navy":                       0x000080, ///< @htmlcolorblock{000080}
		"OldLace":                    0xFDF5E6, ///< @htmlcolorblock{FDF5E6}
		"Olive":                      0x808000, ///< @htmlcolorblock{808000}
		"OliveDrab":                  0x6B8E23, ///< @htmlcolorblock{6B8E23}
		"Orange":                     0xFFA500, ///< @htmlcolorblock{FFA500}
		"OrangeRed":                  0xFF4500, ///< @htmlcolorblock{FF4500}
		"Orchid":                     0xDA70D6, ///< @htmlcolorblock{DA70D6}
		"PaleGoldenrod":              0xEEE8AA, ///< @htmlcolorblock{EEE8AA}
		"PaleGreen":                  0x98FB98, ///< @htmlcolorblock{98FB98}
		"PaleTurquoise":              0xAFEEEE, ///< @htmlcolorblock{AFEEEE}
		"PaleVioletRed":              0xDB7093, ///< @htmlcolorblock{DB7093}
		"PapayaWhip":                 0xFFEFD5, ///< @htmlcolorblock{FFEFD5}
		"PeachPuff":                  0xFFDAB9, ///< @htmlcolorblock{FFDAB9}
		"Peru":                       0xCD853F, ///< @htmlcolorblock{CD853F}
		"Pink":                       0xFFC0CB, ///< @htmlcolorblock{FFC0CB}
		"Plaid":                      0xCC5533, ///< @htmlcolorblock{CC5533}
		"Plum":                       0xDDA0DD, ///< @htmlcolorblock{DDA0DD}
		"PowderBlue":                 0xB0E0E6, ///< @htmlcolorblock{B0E0E6}
		"Purple":                     0x800080, ///< @htmlcolorblock{800080}
		"Red":                        0xFF0000, ///< @htmlcolorblock{FF0000}
		"RosyBrown":                  0xBC8F8F, ///< @htmlcolorblock{BC8F8F}
		"RoyalBlue":                  0x4169E1, ///< @htmlcolorblock{4169E1}
		"SaddleBrown":                0x8B4513, ///< @htmlcolorblock{8B4513}
		"Salmon":                     0xFA8072, ///< @htmlcolorblock{FA8072}
		"SandyBrown":                 0xF4A460, ///< @htmlcolorblock{F4A460}
		"SeaGreen":                   0x2E8B57, ///< @htmlcolorblock{2E8B57}
		"Seashell":                   0xFFF5EE, ///< @htmlcolorblock{FFF5EE}
		"Sienna":                     0xA0522D, ///< @htmlcolorblock{A0522D}
		"Silver":                     0xC0C0C0, ///< @htmlcolorblock{C0C0C0}
		"SkyBlue":                    0x87CEEB, ///< @htmlcolorblock{87CEEB}
		"SlateBlue":                  0x6A5ACD, ///< @htmlcolorblock{6A5ACD}
		"SlateGray":                  0x708090, ///< @htmlcolorblock{708090}
		"SlateGrey":                  0x708090, ///< @htmlcolorblock{708090}
		"Snow":                       0xFFFAFA, ///< @htmlcolorblock{FFFAFA}
		"SpringGreen":                0x00FF7F, ///< @htmlcolorblock{00FF7F}
		"SteelBlue":                  0x4682B4, ///< @htmlcolorblock{4682B4}
		"Tan":                        0xD2B48C, ///< @htmlcolorblock{D2B48C}
		"Teal":                       0x008080, ///< @htmlcolorblock{008080}
		"Thistle":                    0xD8BFD8, ///< @htmlcolorblock{D8BFD8}
		"Tomato":                     0xFF6347, ///< @htmlcolorblock{FF6347}
		"Turquoise":                  0x40E0D0, ///< @htmlcolorblock{40E0D0}
		"Violet":                     0xEE82EE, ///< @htmlcolorblock{EE82EE}
		"Wheat":                      0xF5DEB3, ///< @htmlcolorblock{F5DEB3}
		"White":                      0xFFFFFF, ///< @htmlcolorblock{FFFFFF}
		"WhiteSmoke":                 0xF5F5F5, ///< @htmlcolorblock{F5F5F5}
		"Yellow":                     0xFFFF00, ///< @htmlcolorblock{FFFF00}
		"YellowGreen":                0x9ACD32, ///< @htmlcolorblock{9ACD32}
	}

	// Clients is a map of mac addresses to client objects.
	// All IDs should be lowercase, use Clients.Search() to find a client by ID.
	Clients = map[string]*Client{
		"8c:aa:b5:7a:7d:13": &Client{
			Name:    "Test Client",
			Loc:     TEST,
			NumLEDS: ledCount,
			StepLen: 100,
		},
		"8c:aa:b5:7a:bc:ad": &Client{
			Name:    "Gutter Kitchen",
			Loc:     GUTTER,
			NumLEDS: ledCount,
			StepLen: 100,
		},
		// Placeholder MAC address for the gutter towards the TV room
		"8c:aa:b5:7a:7d:15": &Client{
			Name:    "Gutter TV Room",
			Loc:     GUTTER,
			NumLEDS: ledCount,
			StepLen: 100,
		},
	}
)

type RGBTimeRequest struct {
	Steps []struct {
		// Color is the color to set the LED strip to.
		Color int `json:"color"`
		// Time is the time in milliseconds to transition to the new color.
		Time int `json:"time"`
	}
}

type HSVTimeRequest struct {
	Steps []struct {
		// Color is the color to set the LED strip to.
		Color struct {
			H int `json:"h"`
			S int `json:"s"`
			V int `json:"v"`
		} `json:"color"`
		// Time is the time in milliseconds to transition to the new color.
		Time int `json:"time"`
	}
}

// Search searches the ClientMap for a client with the given name.
// The name here is case-insensitive.
func clientSearch(name string, c map[string]*Client) (*Client, bool) {
	out, ok := c[strings.ToLower(name)]
	return out, ok
}

type Client struct {
	// Name is a user friendly name for the client
	Name string
	// Loc is the location of the client
	Loc location
	// CurrentColor is the current color of the client
	CurrentColor *Resp
	// NumLEDS is the number of LEDs in the client.
	// Standard format is the density of the LED strip * length
	NumLEDS int
	// CurrentColorJSON is the current color of the client marshalled into JSON format.
	// This is exactly what is returned to the client when they request a status update.
	CurrentColorJSON *string
	// StepLen is the time for each step in ms.
	StepLen int
}

func (c *Client) SetColor(cElem *[]ColorElement, ts int64) error {
	c.CurrentColor.Data = cElem
	c.CurrentColor.TS = time.Now().UnixNano()
	if ts != 0 {
		c.CurrentColor.TS = ts
	}
	jsonOut, err := json.Marshal(c.CurrentColor)
	if err != nil {
		return fmt.Errorf("failed to marshal color: %v", err)
	}
	jsonOutStr := string(jsonOut)
	c.CurrentColorJSON = &jsonOutStr
	return nil
}

// RGBColor is an int representing a color in RGB format.
// Example: 0x00FF00 is green.
type RGBColor int

type location int

// A json object to use in responding to the led strip fleet.
// This should be scaled to the length of the light string, based on the request content.
// Set this with c.SetColor()
type Resp struct {
	TS   int64           // Last update time at server.
	Data *[]ColorElement // list of steps to display and color array at this step
}

// As the Clients expect RGB formatted colors, you must convert to RGB before sending.
type Colors []RGBColor

// ColorElement is a single instruction in the color change.
type ColorElement struct {
	Steps  int     // Number of timesteps on client to display the colors.
	Colors *Colors // color list, one per led in the string.
}

func returnAllOneColor(color RGBColor, numLEDS int) *Colors {
	cs := Colors{}
	for i := 0; i < numLEDS; i++ {
		cs = append(cs, color)
	}
	return &cs
}

// handler is the base struct used to handle http services.
type handler struct {
	dictate   Resp
	colorKeys []string
	timestamp time.Time
	port      int
	clients   map[string]*Client
}

func newHandler(port int) (*handler, error) {
	rand.Seed(time.Now().UnixNano())
	colors := []string{}
	for c := range colorDictates {
		colors = append(colors, c)
	}

	return &handler{
		dictate:   Resp{},
		timestamp: time.Now(),
		colorKeys: colors,
		port:      port,
		clients:   Clients,
	}, nil
}

func (h *handler) clientIdleUpdate(t time.Duration) {
	for {
		for _, client := range h.clients {
			if time.Now().UnixNano()-client.CurrentColor.TS > t.Nanoseconds() {
				log.Infof("Setting new color for client: %s", client.Name)
				if err := client.SetColor(h.pickDictate(client), 0); err != nil {
					log.Errorf("failed to SetColor for client %s: %v", client.Name, err)
				}
			}
		}
	}
}

// status returns the current timestamped color dictate to client LED entities.
// Response to GET /status
func (h *handler) status(w http.ResponseWriter, r *http.Request) {
	log.Info("Got status request")

	// Process the variables from the request.
	// id is the MAC address of the client.
	id := r.URL.Query().Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "no client ID")
		return
	}

	// Get the led count from the request.
	ledStr := r.URL.Query().Get("leds")
	stepLenStr := r.URL.Query().Get("len")
	leds, err := strconv.Atoi(ledStr)
	if err != nil {
		log.Errorf("failed to parse ledStr(%s) to int: %v", ledStr, err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "invalid led count")
		return
	}

	// stepLen is the length of time per step in ms.
	stepLen, err := strconv.Atoi(stepLenStr)
	if err != nil {
		log.Errorf("failed to parse stepLenStr(%s) to int: %v", stepLenStr, err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "invalid steplen")
		return
	}

	client, ok := clientSearch(id, h.clients)
	if !ok {
		log.Errorf("unknown client id: %s", id)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "unkonwn client id")
		return
	}
	// Reset the client numLEDS if the client sends that along.
	client.NumLEDS = leds

	log.Infof("Request from client: %s id: %s with stepLen: %d", client.Name, id, stepLen)

	fmt.Fprint(w, *client.CurrentColorJSON)
}

// update handles setting the current value for timestamp and color dictate.
// Response to POST /update
func (h *handler) update(w http.ResponseWriter, r *http.Request) {
	log.Info("Got update request")
	reqUrlSplit := strings.Split(r.URL.Path, "/")

	if len(reqUrlSplit) < 3 {
		log.Errorf("invalid url: %s", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// switch on the second part of the url to determine the update type.
	switch reqUrlSplit[2] {
	case "basic":
		// Sets random color
		h.updateBasic(w, r, reqUrlSplit)
	case "rgbtime":
		// Sets color from JSON response, using hex color values and time in ms.
		h.updateRGBTime(w, r, reqUrlSplit)
	case "hsvtime":
		// Sets color from JSON response, using HSV color values and time in ms.
		h.updateHSVTime(w, r, reqUrlSplit)
	default:
		log.Errorf("invalid url: %s", r.URL.Path)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

// updateHSVTime takes an incoming request and updates the current color element and steps.
func (h *handler) updateHSVTime(w http.ResponseWriter, r *http.Request, reqURLSplit []string) {
	if len(reqURLSplit) != 4 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid url: %s", r.URL.Path)
		return
	}

	// Get the client id from the url.
	id := reqURLSplit[3]
	var ok bool
	client, ok := clientSearch(id, h.clients)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "unknown client id: %s", r.URL.Path)
		return
	}

	// Read the request body.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("failed to read request body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Unmarshal the request body into a color dictate.
	var req HSVTimeRequest
	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Errorf("failed to unmarshal request body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Convert request to color element
	var colors []ColorElement
	for _, step := range req.Steps {
		// convert time to steps
		steps := step.Time / client.StepLen

		// convert HSV to RGB
		r, g, b := colorful.Hsv(float64(step.Color.H), float64(step.Color.S), float64(step.Color.V)).RGB255()
		colors = append(colors, ColorElement{
			Colors: returnAllOneColor(RGBColor((int)(r)<<16|(int)(g)<<8|(int)(b)), client.NumLEDS),
			Steps:  steps,
		})
	}

	err = client.SetColor(&colors, time.Now().UnixNano())
	if err != nil {
		log.Errorf("failed to set color: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed to set color: %v", err)
		return
	}
}

func (h *handler) updateRGBTime(w http.ResponseWriter, r *http.Request, reqURLSplit []string) {
	if len(reqURLSplit) != 4 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid url: %s", r.URL.Path)
		return
	}

	// Get the client id from the url.
	id := reqURLSplit[3]
	var ok bool
	client, ok := clientSearch(id, h.clients)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "unknown client id: %s", r.URL.Path)
		return
	}

	// Read the request body.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("failed to read request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed to read request body: %v", err)
		return
	}

	// Unmarshal the request body into the
	var reqColors RGBTimeRequest
	err = json.Unmarshal(body, &reqColors)
	if err != nil {
		log.Errorf("failed to unmarshal request body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed to unmarshal request body: %v", err)
		return
	}

	// Convert request into a color element
	var colors []ColorElement
	for _, c := range reqColors.Steps {
		// convert time to steps
		steps := c.Time / client.StepLen

		colors = append(colors, ColorElement{
			Colors: returnAllOneColor(RGBColor(c.Color), client.NumLEDS),
			Steps:  steps,
		})
	}

	err = client.SetColor(&colors, time.Now().UnixNano())
	if err != nil {
		log.Errorf("failed to set color: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed to set color: %v", err)
		return
	}
}

// updateBasic picks a random color to select from and applies it statically to the client.
func (h *handler) updateBasic(w http.ResponseWriter, r *http.Request, reqUrlSplit []string) {
	if len(reqUrlSplit) != 4 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "invalid url: %s", r.URL.Path)
		return
	}

	// Get the client id from the url.
	id := reqUrlSplit[3]
	var ok bool
	client, ok := clientSearch(id, h.clients)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "unknown client id: %s", r.URL.Path)
		return
	}

	// Pick a random color to dictate to the client.
	if err := client.SetColor(h.pickDictate(client), 0); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "failed to SetColor for client: %s", r.URL.Path)
	}
	log.Infof("Updated client: %s id: %s with color: %v", client.Name, id, *client.CurrentColor.Data)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "success SetColor")
}

func (h *handler) pickDictate(client *Client) *[]ColorElement {
	// Get a single color randomly from the colorDictates map.
	color := colorDictates[h.colorKeys[rand.Intn(len(h.colorKeys))]]

	return &[]ColorElement{
		{
			Steps:  1,
			Colors: returnAllOneColor(color, client.NumLEDS),
		},
	}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Infof("Got request for: %v", r.RequestURI)
	switch {
	case strings.HasPrefix(r.URL.Path, "/update"):
		h.update(w, r)
	case strings.HasPrefix(r.URL.Path, "/status"):
		h.status(w, r)
	case r.URL.Path == "/":
		http.ServeFile(w, r, "src/index.html")
	case strings.HasPrefix(r.URL.Path, "/static"):
		// Serve static files.
		http.StripPrefix("/static/", http.FileServer(http.Dir("src/static"))).ServeHTTP(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func initClients(leds int) {
	// Define all clients to have a default color dictate.
	// White is used to test all LEDs quickly.
	for _, c := range Clients {
		c.CurrentColor = &Resp{}
		c.NumLEDS = leds
		if err := c.SetColor(&[]ColorElement{{
			Steps:  1,
			Colors: returnAllOneColor(0xFFFFFF, c.NumLEDS)},
		}, 1); err != nil {
			log.Errorf("failed to SetColor for client: %s: %v", c.Name, err)
		}
	}
}

func main() {
	flag.Parse()
	log.Infof("Server will listen on port: %d", *port)

	h, err := newHandler(*port)
	if err != nil {
		log.Fatalf("failed to create handler: %v", err)
	}

	// Initialize the client data structures.
	initClients(ledCount)

	go h.clientIdleUpdate(idleTime)

	// Start a goroutine that will force a change

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", *host, *port),
		Handler: h,
	}
	log.Fatal(s.ListenAndServe())
}
