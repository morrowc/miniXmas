// Run on an ESP device, control an LED (ws2812-type) string.
// Collect color settings to implement from an HTTP request to a hardcoded
// URL.
#include <string.h>
#include <Arduino.h>
#include <ArduinoJson.h>
// #include "FastLED.h"
#include <ESP8266WiFi.h>
#include <ESP8266WiFiMulti.h>
#include <ESP8266HTTPClient.h>
#include <WiFiClient.h>

/*
#if FASTLED_VERSION < 3001000
#error "Requires FastLED 3.1 or later; check github for latest code."
#endif
*/

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
const unsigned long STEP_DELAY = 100;
// Send the light back/forth like a cylon.
const int CYLON_DELAY = 100;
// For how many cylces to be a cylon?
const int CYCLES = 10;
// The delimiter between reply parts from the controller.
const char* DELIMITER = ", ";
// The current timestamp value from the previous controller reply.
long long CURRENT = 0;
// The Array of LEDs to control.
// CRGB leds[NUM_LEDS];

// From github.com/bigjosh/SimpleNeoPixelDemo/SimpleNeoPixelTimingTester
// These values are for digital pin 8 on an Arduino Yun or digital pin 12 on a DueMilinove
// Note that you could also include the DigitalWriteFast header file to not need to to this lookup.

#define PIXEL_DDR   1   // Port of the pin the pixels are connected to
#define PIXEL_BIT   4      // Bit of the pin the pixels are connected to

// These are the timing constraints taken mostly from the WS2812 datasheet 

#define T1H  700    // Width of a 1 bit in ns
#define T1L  600    // Width of a 1 bit in ns

#define T0H  350    // Width of a 0 bit in ns
#define T0L  800    // Width of a 0 bit in ns

#define RES 6000    // Width of the low gap between bits to cause a frame to latch

// Here are some convience defines for using nanoseconds specs to generate actual CPU delays

// Note that this has to be SIGNED since we want to be able to check for negative values of derivatives
#define NS_PER_SEC (1000000000L)        

#define CYCLES_PER_SEC (F_CPU)

#define NS_PER_CYCLE ( NS_PER_SEC / CYCLES_PER_SEC )

#define NS_TO_CYCLES(n) ( (n) / NS_PER_CYCLE )

// Make sure we never have a delay less than zero
#define DELAY_CYCLES(n) ( ((n)>0) ? __builtin_avr_delay_cycles( n ) :  __builtin_avr_delay_cycles( 0 ) ) 

void  sendBit(bool) __attribute__ ((optimize(0)));

void sendBit( bool bitVal ) {
    if (  bitVal ) {
      // bitSet(DATA_PIN, PIXEL_BIT);
      // digitalWrite(DATA_PIN, bitVal);
      digitalWrite(DATA_PIN, (bitVal << 2));
      // 1-bit width less  overhead  for the actual bit setting
      // DELAY_CYCLES(NS_TO_CYCLES(T1H) - 2);
      delayMicroseconds(700);
      // bitClear(DATA_PIN, PIXEL_BIT);
      digitalWrite(DATA_PIN, 0);
      // 1-bit gap less the overhead of the loop
      // DELAY_CYCLES(NS_TO_CYCLES(T1L) - 10);
      delayMicroseconds(700);
    } else {
      // 0-bit width less overhead
      // bitSet(DATA_PIN, PIXEL_BIT);
      digitalWrite(DATA_PIN, bitVal);
      // DELAY_CYCLES(NS_TO_CYCLES(T0H) - 2);
      delayMicroseconds(350);
      // **************************************************************************
      // This line is really the only tight goldilocks timing in the whole program!
      // **************************************************************************
      // bitClear(DATA_PIN, PIXEL_BIT);
      digitalWrite(DATA_PIN, 0);
      // 0-bit gap less overhead of the loop
      // DELAY_CYCLES(NS_TO_CYCLES(T0L) - 10);
      delayMicroseconds(600);
    }
    /*
    * Note that the inter-bit gap can be as long as you want as long as it
    * doesn't exceed the 5us reset timeout (which is A long time). Here I
    * have been generous and not tried to squeeze the gap tight but instead
    * erred on the side of lots of extra time. This has thenice side effect
    * of avoid glitches on very long strings becuase 
    */
}  

void sendByte( unsigned char byte ) {

    for( unsigned char bit = 0 ; bit < 8 ; bit++ ) {

      sendBit( bitRead( byte , 7 ) );                // Neopixel wants bit in highest-to-lowest order
                                                     // so send highest bit (bit #7 in an 8-bit byte since they start at 0)
      byte <<= 1;                                    // and then shift left so bit 6 moves into 7, 5 moves into 6, etc

    }
}

/*
  The following three functions are the public API:
  
  ledSetup() - set up the pin that is connected to the string. Call once at the begining of the program.  
  sendPixel( r g , b ) - send a single pixel to the string. Call this once for each pixel in a frame.
  show() - show the recently sent pixel on the LEDs . Call once per frame. 
  
*/


// Set the specified pin up as digital out

void ledsetup() {
  
  // bitSet( PIXEL_DDR , PIXEL_BIT );
  digitalWrite(DATA_PIN, 1);
  
}

void sendPixel( unsigned char r, unsigned char g , unsigned char b )  {

  sendByte(g);          // Neopixel wants colors in green then red then blue order
  sendByte(r);
  sendByte(b);

}


// Just wait long enough without sending any bots to cause the pixels to latch and display the last sent frame

void show() {
    // DELAY_CYCLES( NS_TO_CYCLES(RES) );
      delayMicroseconds(600);
}

void showColor( unsigned char r , unsigned char g , unsigned char b ) {

  cli();
  for( int p=0; p<NUM_LEDS; p++ ) {
    sendPixel( r , g , b );
  }
  sei();
  show();

}

void setup() {
  pinMode(DATA_PIN , OUTPUT);
  Serial.begin(9600);
  // ledsetup();
  
}

// Simple blink on/off.
void loop() {
  Serial.println("Sending zeros");
  os_intr_lock();
  showColor(0, 0, 0);
  delay(500);
  Serial.println("Sending red");
  showColor(0xff, 0x00, 0x00);
  delay(500);
  os_intr_unlock();
}







//////////////////////////////////// break //////////////////

/*
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
  if ((WiFi.status() == WL_CONNECTED)) {}
  
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

void checkDelay(int st) {
  while ( (millis() - st) < DELAY ) {}
}

void loop()
{
  int st = millis();
  Serial.printf("Millis: %d", st);
  Serial.println();
  char url[strlen(URL)+50];
  sprintf(url, "%s?id=%s&leds=%d&len=%d", URL, WiFi.macAddress().c_str(), NUM_LEDS, DELAY);

  String  payload = doHttp(url);
  // Determine how long the payload is, if zero, error and return..
  if (payload.length() == 0) {
    Serial.println("Got zero length HTTP reply");
    Serial.println();
    checkDelay(st);
    return;
  }

  // Create a JSON Document, and deserialize payload into that.
  // NOTE: the 10k used here is a guestimate.
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


  // Handle each step, with a request to the HTTP service at
  // each step start.
  /*
   * Example data.
  {
   "TS":1670778488327762396,
   "Data":[
      {
         "Steps":2,
         "Colors":[
            10145074,
            10145074,
            10145074
            ]
       },
      {
         "Steps":2,
         "Colors":[
            1231234,
            1231234,
            1231234
            ]
       }
    ]
  }
  */
  /*
  size_t arr_size = doc["Data"].size();
  Serial.println();
  for (int s = 0; s < arr_size; s++) {
    StaticJsonDocument<10000> data = doc["Data"][s];
    // Collect the period of time to set the intended color.
    String steps = data["Steps"];
    int stepsInt = steps.toInt();

    // Loop the number of steps, with DELAY in between.
    for (int l = 0; l < stepsInt; l++) {
      for (int i = 0; i < NUM_LEDS; i++) {
          String color = data["Colors"][i];
          int cInt = color.toInt();
          leds[i] = cInt;
          FastLED.show();
      }
      delay(STEP_DELAY);
    }
  }
  // Delay until after the reuqired wait period between changes ocurs.
  checkDelay(st);
}

void bot_to_top(StaticJsonDocument<1000> data, int cDelay) {
    // Chase 5 at a time down the pipe.
    for (int i = 0; i < NUM_LEDS; i += 5) {
        String color = data["Colors"][i];
        int cInt = color.toInt();
        leds[i] = cInt;
        leds[i+1] = cInt;
        leds[i+2] = cInt;
        leds[i+3] = cInt;
        leds[i+4] = cInt;
        FastLED.show();
        leds[i] = 0;
        leds[i+1] = 0;
        leds[i+2] = 0;
        leds[i+3] = 0;
        leds[i+4] = 0;
        delay(cDelay);
   }
}

void top_to_bot(StaticJsonDocument<1000> data, int cDelay) {
    // Chase 5 at a time down the pipe.
    for (int i = NUM_LEDS; i > 0; i -= 5) {
        String color = data["Colors"][i];
        int cInt = color.toInt();
        leds[i] = cInt;
        leds[i-1] = cInt;
        leds[i-2] = cInt;
        leds[i-3] = cInt;
        leds[i-4] = cInt;
        FastLED.show();
        leds[i] = 0;
        leds[i-1] = 0;
        leds[i-2] = 0;
        leds[i-3] = 0;
        leds[i-4] = 0;
        delay(cDelay);
   }
}
*/
