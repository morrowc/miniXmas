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
#define NUM_LEDS    1
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
#define PIXEL_BIT   B11   // Bit of the pin the pixels are connected to

// These are the timing constraints taken mostly from the WS2812 datasheet 

#define T1H  700    // Width of a 1 bit in ns
#define T1L  600    // Width of a 1 bit in ns

#define T0H  350    // Width of a 0 bit in ns
#define T0L  800    // Width of a 0 bit in ns

#define RES 6000    // Width of the low gap between bits to cause a frame to latch
#define LOOP_DELAY 41

// Here are some convience defines for using nanoseconds specs to generate actual CPU delays

// Note that this has to be SIGNED since we want to be able to check for negative values of derivatives
#define NS_PER_SEC (1000000000L)        

#define CYCLES_PER_SEC (F_CPU)

#define NS_PER_CYCLE ( NS_PER_SEC / CYCLES_PER_SEC )

#define NS_TO_CYCLES(n) ( (n) / NS_PER_CYCLE )

// Make sure we never have a delay less than zero
// #define DELAY_CYCLES(n) ( ((n)>0) ? __builtin_avr_delay_cycles( n ) :  __builtin_avr_delay_cycles( 0 ) ) 

#define PIN4_MUX PERIPHS_IO_MUX_GPIO4_U
#define PIN4_FUNC FUNC_GPIO4
#define PIN4_PIN 4

void  sendBit(bool) __attribute__ ((optimize(0)));

void sendBit( bool bitVal ) {
    if (  bitVal ) {  // 0 bit
      // Turn the bit on.
      digitalWrite(DATA_PIN, HIGH);
      /*
      asm volatile (
			".rept %[onCycles] \n\t"          // Execute NOPs to delay exactly the specified number of cycles
			"nop \n\t"
			".endr \n\t"
      ::
			[onCycles]	"I"  ((NS_TO_CYCLES(T1H) - 2) / LOOP_DELAY)		// 1-bit width less overhead  for the actual bit
                                                // setting, note that this delay could be longer
                                                // and everything would still work
		  );
      */

      // Turn the bit off.
      digitalWrite(DATA_PIN, LOW);
      /*
      asm volatile (
			".rept %[offCycles] \n\t"         // Execute NOPs to delay exactly the specified number of cycles
			"nop \n\t"
			".endr \n\t"
			::
			[offCycles] 	"I" ((NS_TO_CYCLES(T1L) - 2) / LOOP_DELAY)		// Minimum interbit delay. Note that we probably
                                                  // don't need this at all since the loop overhead
                                                  // will be enough, but here for correctness
		  );
      */
    } else {
      // **************************************************************************
		  // This line is really the only tight goldilocks timing in the whole program!
		  // **************************************************************************
      digitalWrite(DATA_PIN, HIGH);
		  /*asm volatile (
		  	".rept %[onCycles] \n\t"				  // Now timing actually matters. The 0-bit must be long
                                          // enough to be detected but not too long or it will be a 1-bit
		  	"nop \n\t"                        // Execute NOPs to delay exactly the specified number of cycles
		  	".endr \n\t"
		  	::
		  	[onCycles]	"I" ((NS_TO_CYCLES(T0H) - 2) / LOOP_DELAY)
		  );
      */

      digitalWrite(DATA_PIN, LOW);
      /*
      asm volatile (
		  	".rept %[offCycles] \n\t"         // Execute NOPs to delay exactly the specified number of cycles
		  	"nop \n\t"
		  	".endr \n\t"
		  	::
		  	[offCycles]	"I" ((NS_TO_CYCLES(T0L) - 2) / LOOP_DELAY)
		  );
      */
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

      sendBit( bitRead( byte , 7 ) ); // Neopixel wants bit in highest-to-lowest order
                                      // so send highest bit (bit #7 in an 8-bit byte since they start at 0)
      byte <<= 1;                     // and then shift left so bit 6 moves into 7, 5 moves into 6, etc

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
  
  // WRITE_PERI_REG( PERIPHS_GPIO_BASEADDR + 4, PIXEL_BIT );
  // bitSet( PIXEL_DDR , PIXEL_BIT );
  // digitalWrite(DATA_PIN, 1);
  
}

void sendPixel( unsigned char r, unsigned char g , unsigned char b )  {
  // Neopixel wants colors in green then red then blue order
  sendByte(g);
  sendByte(r);
  sendByte(b);
}


// Just wait long enough without sending any bots to cause the
// pixels to latch and display the last sent frame
void show() {
    // DELAY_CYCLES( NS_TO_CYCLES(RES) );
      delayMicroseconds(600);
}

void showColor( unsigned char r , unsigned char g , unsigned char b ) {
  // clear interrupts.
  cli();
  for( int p=0; p<NUM_LEDS; p++ ) {
    sendPixel( r , g , b );
  }
  // Restart interrupts.
  sei();
  show();
}

void setup() {
  pinMode(DATA_PIN , OUTPUT);
  PIN_FUNC_SELECT(PIN4_MUX, PIN4_FUNC);
  PIN_PULLUP_EN(PIN4_MUX);
  Serial.begin(9600);
  // ledsetup();
}

// Simple blink on/off.
void loop() {
  // Serial.println("Sending zeros");
  os_intr_lock();
  /*
  for (int i=0; i <= 256; i += 16) {
    Serial.printf("Color %d\n", i);
    Serial.println();
    showColor(i, i, i);
    delay(500);
  }
  */
  showColor(128, 128, 128);
  delay(500);
  os_intr_unlock();
}
