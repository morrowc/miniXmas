// Run on an ESP device, control an LED (ws2812-type) string.
// Collect color settings to implement from an HTTP request to a hardcoded
// URL.
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

const char* HOST = "mailserver.ops-netman.net";
const uint16 PORT = 6789;
const char* URI = "/status";
const char* SSID = "theaternet";
const char* PASS = "network123";

CRGB leds[NUM_LEDS];


void setup() {
  // Setup the serial output for console/logging.
  Serial.begin(9600);
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
  if (!client.connect(HOST, PORT)) {
    Serial.println("Connection failed");
    delay(5000);
  }
  Serial.println("Sending request to server");
  if (client.connected()) {
    client.println(URI);
  }
  // Wait until the client returns from getting all of the data.
  unsigned long timeout = millis();
  while (client.available() == 0) {
    if (millis() - timeout > 5000) {
      Serial.println(">>> Client timeout !");
      client.stop();
      delay(5000);
      break;
    }
  }

  // Read the data from the client buffer.
  Serial.println("Receiving data from server.");
  while (client.available()) {
    char ch = client.read();
    Serial.print(ch);
  }

  // Report and close the connection.
  Serial.println();
  Serial.println("closing the connection");
  client.stop();

  pride();
  FastLED.show();  
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

