package nuclearrng

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
)

type NuclearRNG struct {
	sync.Mutex
	port serial.Port
}

func New() (*NuclearRNG, error) {
	portsDetails, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, fmt.Errorf("nuclearrng: %w", err)
	}

	var portName string
	for _, portDetails := range portsDetails {
		if portDetails.IsUSB && validateVIDPID(portDetails.VID, portDetails.PID) {
			portName = portDetails.Name
			break
		}
	}
	if portName == "" {
		return nil, errors.New("nuclearrng: no attached device found")
	}

	port, err := serial.Open(portName, &serial.Mode{BaudRate: 115200})
	if err != nil {
		return nil, fmt.Errorf("nuclearrng: %w", err)
	}

	return &NuclearRNG{port: port}, nil
}

func NewByVIDPID(vid, pid string) (*NuclearRNG, error) {
	portsDetails, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, fmt.Errorf("nuclearrng: %w", err)
	}

	var portName string
	for _, portDetails := range portsDetails {
		if portDetails.IsUSB && portDetails.VID == vid && portDetails.PID == pid {
			portName = portDetails.Name
			break
		}
	}
	if portName == "" {
		return nil, fmt.Errorf("nuclearrng: no attached device found with VID/PID %04x/%04x", vid, pid)
	}

	port, err := serial.Open(portName, &serial.Mode{BaudRate: 115200})
	if err != nil {
		return nil, fmt.Errorf("nuclearrng: %w", err)
	}

	return &NuclearRNG{port: port}, nil
}

func (rng *NuclearRNG) Read(p []byte) (n int, err error) {
	rng.Lock()
	defer rng.Unlock()
	return rng.port.Read(p)
}

func (rng *NuclearRNG) Close() error {
	rng.Lock()
	defer rng.Unlock()
	return rng.port.Close()
}

func validateVIDPID(vid, pid string) bool {
	vidpid := fmt.Sprintf("%s:%s", vid, pid)
	for _, validVIDPID := range validVIDPIDs {
		if strings.ToLower(vidpid) == strings.ToLower(validVIDPID) {
			return true
		}
	}
	return false
}

var validVIDPIDs = []string{
	"2e8a:000a", // RPi Pico
}
