package nuclearrng

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"go.bug.st/serial"
	"go.bug.st/serial/enumerator"
	"golang.org/x/crypto/chacha20"
)

const rekeyBase = 1024 * 1024

type NuclearRNG struct {
	sync.Mutex
	port serial.Port

	cipher        *chacha20.Cipher
	reseedCounter int
}

func New() (*NuclearRNG, error) {
	rng, err := NewRaw()
	if err != nil {
		return nil, err
	}

	buf := make([]byte, chacha20.KeySize+chacha20.NonceSize)
	_, err = rng.readRaw(buf)
	if err != nil {
		return nil, err
	}
	rng.cipher, err = chacha20.NewUnauthenticatedCipher(buf[:chacha20.KeySize], buf[chacha20.KeySize:])
	if err != nil {
		return nil, fmt.Errorf("nuclearrng: %w", err)
	}

	return rng, nil
}

func NewRaw() (*NuclearRNG, error) {
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

func (rng *NuclearRNG) stirIfNeeded(length int) error {
	if rng.reseedCounter <= length {
		buf := make([]byte, chacha20.KeySize+chacha20.NonceSize)
		_, err := rng.readRaw(buf)
		if err != nil {
			return err
		}

		err = rng.rekey(buf)
		if err != nil {
			return err
		}

		fuzzBytes := make([]byte, 64/8)
		rng.cipher.XORKeyStream(fuzzBytes, fuzzBytes)
		rng.reseedCounter = rekeyBase + int(binary.BigEndian.Uint64(fuzzBytes)%uint64(rekeyBase))
	}

	if rng.reseedCounter <= length {
		rng.reseedCounter = 0
	} else {
		rng.reseedCounter -= length
	}

	return nil
}

func (rng *NuclearRNG) rekey(data []byte) (err error) {
	buf := make([]byte, chacha20.KeySize+chacha20.NonceSize)
	copy(buf, data)
	rng.cipher.XORKeyStream(buf, buf)
	rng.cipher, err = chacha20.NewUnauthenticatedCipher(buf[:chacha20.KeySize], buf[chacha20.KeySize:])
	if err != nil {
		return fmt.Errorf("nuclearrng: %w", err)
	}
	return nil
}

func (rng *NuclearRNG) GetRandom(bytes int) ([]byte, error) {
	ret := make([]byte, bytes)
	_, err := rng.Read(ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (rng *NuclearRNG) Read(p []byte) (n int, err error) {
	rng.Lock()
	defer rng.Unlock()

	if rng.cipher == nil {
		return rng.readRaw(p)
	}

	rng.stirIfNeeded(len(p))
	rng.cipher.XORKeyStream(p, p)

	return len(p), nil
}

func (rng *NuclearRNG) readRaw(p []byte) (int, error) {
	count := 0
	for count < len(p) {
		n, err := rng.port.Read(p[count:])
		if err != nil {
			return n, fmt.Errorf("nuclearrng: %w", err)
		}
		count += n
	}
	log.Printf("nuclearrng: readRaw() %s", hex.EncodeToString(p))
	return count, nil
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
