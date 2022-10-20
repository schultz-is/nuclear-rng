package main

import (
	"fmt"
	"machine"
	"time"

	"github.com/schultz-is/nuclear-rng/firmware/board"
)

func main() {
	// Configure signal pin as input and pull up
	board.SignalPin.Configure(machine.PinConfig{Mode: machine.PinInputPullup})

	// Configure on-board LED pin as output
	board.LEDPin.Configure(machine.PinConfig{Mode: machine.PinOutput})

	var tick bool            // Tick cycle status for bias mitigation
	var led bool             // LED on/off status
	var sig1, sig2 bool      // Signal status to track blip falling edge
	var t1, t2, t3 time.Time // Signal blip timestamps
	var d1, d2 time.Duration // Signal time deltas
	var buf byte             // RNG buffer byte
	var bufIdx uint8         // RNG buffer bit index

	for {
		// Slide the signal window
		sig1, sig2 = sig2, board.SignalPin.Get()

		// Continue the main loop unless a falling edge is detected
		if !(sig1 && !sig2) {
			continue
		}

		// Toggle tick
		tick = !tick

		// Slide the time window
		t1, t2, t3 = t2, t3, time.Now()

		// Continue the main loop when fewer than three measurements are present
		if t1.IsZero() {
			continue
		}

		// Calculate timestamp deltas
		d1, d2 = d2, t3.Sub(t2)

		// Shift the RNG buffer left one bit
		buf <<= 1

		// This is a technique to mitigate any bias that might occur due to nuclear decay. The more a
		// material decays, the less active it is, which can lead to d2 being more likely to be longer
		// than d1. This bias is minimal, especially depending on the half-life of measured material,
		// but worth accounting for.
		if tick {
			// Toggle the current bit when d1 > d2 on a tick cycle
			if d1 > d2 {
				buf |= 1
			}
		} else {
			// Toggle the current bit when d2 > d1 on a non-tick cycle
			if d2 > d1 {
				buf |= 1
			}
		}

		// Increment the buffer bit index
		bufIdx++

		// When the RNG buffer is full (one byte of data,) write it out and toggle the on-board LED
		if bufIdx > 7 {
			led = !led
			board.LEDPin.Set(led)
			fmt.Printf("%02x", buf)
			buf, bufIdx = 0, 0
		}
	}
}
