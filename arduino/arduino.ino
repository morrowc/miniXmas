// Run on an ESP device, control an LED (ws2812-type) string.
// Collect color settings to implement from an HTTP request to a hardcoded
// URL.
#include <string.h>
#include <Arduino.h>
#include <ArduinoJson.h>
#include "FastLED.h"
#include <ESP8266WiFi.h>
#include <ESP8266WiFiMulti.h>
#include <ESP8266HTTPClient.h>
#include <WiFiClient.h>

#if FASTLED_VERSION < 3001000
#error "Requires FastLED 3.1 or later; check github for latest code."
#endif

#define DATA_PIN    2
#define LED_TYPE    NEOPIXEL
#define COLOR_ORDER RGB
#define NUM_LEDS    30
#define BRIGHTNESS  200
#define  ARDUINOJSON_USE_LONG_LONG 1

const char* URL = "http://mailserver.ops-netman.net:6789/status";
const char* SSID = "theaternet";
const char* PASS = "network123";
// Delay betwen web requests and possible change to lights.
const unsigned long DELAY = 1000;
// The delimiter between reply parts from the controller.
const char* DELIMITER = ", ";
// The current timestamp value from the previous controller reply.
long long CURRENT = 0;
// The ID of this board, it's MacAddress.
String ID = "";

CRGB leds[NUM_LEDS];
unsigned long TOTAL_INTERVALS = 0;
unsigned long WAIT_TIME = 0;


void setup() {
  // Setup the serial output for console/logging.
  Serial.begin(115200);
  Serial.setDebugOutput(true);
  Serial.println();
  Serial.println("[setup]: Starting up.");

  delay(3000); // 3 second delay for recovery

  // Connect to wifi, as a station.
  WiFi.mode(WIFI_STA);
  WiFi.begin(SSID, PASS);
  delay(5000); // 5 second delay for recovery
  Serial.println("connected to wifi");
  if ((WiFi.status() == WL_CONNECTED)) {
    WiFi.printDiag(Serial);
  }
  
  // Set the ID/mac-addr.
  ID = WiFi.macAddress();

  // tell FastLED about the LED strip configuration
  FastLED.addLeds<LED_TYPE,DATA_PIN>(leds, NUM_LEDS);
  FastLED.setCorrection(TypicalLEDStrip);
  FastLED.setDither(BRIGHTNESS < 255);
  FastLED.setBrightness(BRIGHTNESS);
}

// Handle sending a request to the http server, return the raw payload.
String doHttp(char* url) {
  // Create an http client in this version of the loop, collect data from
  // remote server.
  WiFiClient client;
  HTTPClient http;
  http.begin(client, url);
  int httpResponseCode = http.GET();
  if (httpResponseCode>0) {
    // Get the payload as a String()
    return http.getString();
  }
  return String("");
}

void loop()
{
  int st = millis();
  Serial.printf("Millis: %d", st);
  Serial.println();
  char url[strlen(URL)+50];
  sprintf(url, "%s?id=%s&leds=%d&len=%d", URL, ID.c_str(), NUM_LEDS, DELAY);

  // Set the default dictate to 'rainbow'.
  String DICTATE = "rainbow";
  String  payload = doHttp(url);
  // Determine how long the payload is.
  if (payload.length() == 0) {
    Serial.println("Got zero length HTTP reply");
    Serial.println();
    checkDelay(st);
    return;
  }

  // Create a JSON Document, and deserialize payload into that.
  StaticJsonDocument<10000> doc;

  DeserializationError error = deserializeJson(doc, payload);

  if (error) {
    Serial.print(F("deserializeJson() failed: "));
    Serial.println(error.f_str());
    checkDelay(st);
    return;
  }

  long long TS = doc["TS"]; // 1670710298274952000
  if (CURRENT != TS) {
    CURRENT = TS;
  } else {
    Serial.print("no change in TimeStamp");
    Serial.println();
    checkDelay(st);
    return;
  }

  TOTAL_INTERVALS++;
  WAIT_TIME += (millis() - st);
  Serial.printf("%d / %d == %d", WAIT_TIME, TOTAL_INTERVALS,
    WAIT_TIME / TOTAL_INTERVALS);
  Serial.println();

  // Handle each step, with a request to the HTTP service at
  // each step start.
  for (int i = 0; i < NUM_LEDS; i++) {
      String color = doc["Data"][0]["Colors"][i];
      int cInt = color.toInt();
      leds[i] = cInt;
      // Serial.printf("Color: %s Num: %d", color, cInt);
      FastLED.show();
  }
  // Delay until after the reuqired wait period between changes ocurs.
  checkDelay(st);
}

void checkDelay(int st) {
  while ( (millis() - st) < DELAY ) {}
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

