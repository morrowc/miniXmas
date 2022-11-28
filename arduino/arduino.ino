// Run on an ESP device, control an LED (ws2812-type) string.
// Collect color settings to implement from an HTTP request to a hardcoded
// URL.
#include <string.h>
#include <Arduino.h>
#include "FastLED.h"
#include <ESP8266WiFi.h>
#include <ESP8266WiFiMulti.h>
#include <ESP8266HTTPClient.h>
#include <WiFiClient.h>

#if FASTLED_VERSION < 3001000
#error "Requires FastLED 3.1 or later; check github for latest code."
#endif

#define DATA_PIN    2
// #define CLK_PIN   4
#define LED_TYPE    NEOPIXEL
#define COLOR_ORDER RGB
#define NUM_LEDS    30
#define BRIGHTNESS  100

const char* URL = "http://mailserver.ops-netman.net:6789/status";
const char* SSID = "theaternet";
const char* PASS = "network123";
// The delimiter between reply parts from the controller.
const char* DELIMITER = ", ";
// The current timestamp value from the previous controller reply.
String CURRENT = "";

CRGB leds[NUM_LEDS];


void setup() {
  // Setup the serial output for console/logging.
  Serial.begin(115200);
  Serial.setDebugOutput(true);
  Serial.println();
  Serial.println();
  Serial.println();
  Serial.println("[setup]: Starting up.");

  delay(3000); // 3 second delay for recovery

  // Connect to wifi, as a station.
  WiFi.mode(WIFI_STA);
  WiFi.begin(SSID, PASS);
  Serial.println("connected to wifi");
  if ((WiFi.status() == WL_CONNECTED)) {
    WiFi.printDiag(Serial);
  }
  
  // tell FastLED about the LED strip configuration
  FastLED.addLeds<LED_TYPE,DATA_PIN>(leds, NUM_LEDS)
    .setCorrection(TypicalLEDStrip)
    .setDither(BRIGHTNESS < 255);

  // set master brightness control
  FastLED.setBrightness(BRIGHTNESS);
}

void loop()
{
  // Create an http client in this version of the loop, collect data from
  // remote server.
  WiFiClient client;
  HTTPClient http;
  String url = URL;
  // Set the default dictate to 'rainbow'.
  String DICTATE = "rainbow";

  Serial.println("Starting http client request");
  http.begin(client, url.c_str());
  int httpResponseCode = http.GET();
  if (httpResponseCode>0) {
    Serial.print("HTTP Response Code: ");
    Serial.println(httpResponseCode);
    String payload = http.getString();
    Serial.println(payload);
    // Convert payload from String to char[] and from char[] to char*.
    int n = payload.length();
    char p_array[n + 1];
    strcpy(p_array, payload.c_str());
    char* p;
    for (int i = 0; i < n; i++) {
      p += p_array[i];
    }
    free(p_array);
    // Use strtok() to tokenize the http payload.
    // First token should be the timestamp, second is the dictate.
    String token = strtok(p, DELIMITER);
    if ( token != NULL ) {
      if ( token != CURRENT) {
        CURRENT = token;
        token = strtok(NULL, DELIMITER);
        if ( token != NULL ) { DICTATE = token; }
        Serial.printf("TS: %s Dictate: %s", CURRENT, DICTATE);
        // Display rainbow for now.
        fill_rainbow(leds, NUM_LEDS, 0, 5);
        FastLED.show();
      }
    }
  } else {
    // 
    pride();
    FastLED.show();  
  }
    delay(5000);
}


// This function draws rainbows with an ever-changing,
// widely-varying set of parameters.
void pride() 
{
  static uint16_t sPseudotime = 0;
  static uint16_t sLastMillis = 0;
  static uint16_t sHue16 = 0;
 
  uint8_t sat8 = beatsin88( 87, 220, 250);
  uint8_t brightdepth = beatsin88( 341, 96, 224);
  uint16_t brightnessthetainc16 = beatsin88( 203, (25 * 256), (40 * 256));
  uint8_t msmultiplier = beatsin88(147, 23, 60);

  uint16_t hue16 = sHue16;//gHue * 256;
  uint16_t hueinc16 = beatsin88(113, 1, 3000);
  
  uint16_t ms = millis();
  uint16_t deltams = ms - sLastMillis ;
  sLastMillis  = ms;
  sPseudotime += deltams * msmultiplier;
  sHue16 += deltams * beatsin88( 400, 5,9);
  uint16_t brightnesstheta16 = sPseudotime;
  
  for( uint16_t i = 0 ; i < NUM_LEDS; i++) {
    hue16 += hueinc16;
    uint8_t hue8 = hue16 / 256;

    brightnesstheta16  += brightnessthetainc16;
    uint16_t b16 = sin16( brightnesstheta16  ) + 32768;

    uint16_t bri16 = (uint32_t)((uint32_t)b16 * (uint32_t)b16) / 65536;
    uint8_t bri8 = (uint32_t)(((uint32_t)bri16) * brightdepth) / 65536;
    bri8 += (255 - brightdepth);
    
    CRGB newcolor = CHSV( hue8, sat8, bri8);
    
    uint16_t pixelnumber = i;
    pixelnumber = (NUM_LEDS-1) - pixelnumber;
    
    nblend( leds[pixelnumber], newcolor, 64);
  }
}

