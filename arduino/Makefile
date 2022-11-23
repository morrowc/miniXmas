compile: arduino.ino
	/home/morrowc/bin/arduino-cli compile --fqbn esp8266:esp8266:nodemcuv2 ${PWD}

upload: arduino.ino
	/home/morrowc/bin/arduino-cli upload --fqbn esp8266:esp8266:nodemcuv2 -p /dev/ttyUSB0 ${PWD}
