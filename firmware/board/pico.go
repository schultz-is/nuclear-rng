//go:build pico
// +build pico

package board

import "machine"

const SignalPin machine.Pin = machine.GP0 // Incoming signal pin
const LEDPin machine.Pin = machine.LED    // On-board LED pin
