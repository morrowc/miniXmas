BFLAGS = --build-property build.f_cpu=160000000L

all: compile upload

compile: arduino.ino
	 arduino-cli compile --fqbn esp8266:esp8266:d1_mini ${BFLAGS}  ${PWD}

upload: arduino.ino
	 arduino-cli upload --fqbn esp8266:esp8266:d1_mini -p /dev/ttyUSB0 ${PWD}

