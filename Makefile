GO ?= go
TINYGO ?= tinygo
FIRMWARE_TARGET ?= pico
BUILD_OPTIONS ?= "-size=short"
FLASH_OPTIONS ?= "-size=short"

build/firmware:
	cd firmware && $(TINYGO) build $(BUILD_OPTIONS) -target=$(FIRMWARE_TARGET) firmware.go

flash/firmware:
	cd firmware && $(TINYGO) flash $(FLASH_OPTIONS) -target=$(FIRMWARE_TARGET) firmware.go

monitor/firmware:
	cd firmware && $(TINYGO) monitor -target=$(FIRMWARE_TARGET)
