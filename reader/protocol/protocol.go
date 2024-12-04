package protocol

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/s5i/ruuvi2db/data"
)

// ParseDatagram converts raw BLE datagram to data.Point.
func ParseDatagram(mfID uint16, data []byte, addr string) (dp *data.Point, e error) {
	errs := []string{}
	defer func() {
		if dp != nil {
			dp.Address = strings.ToUpper(addr)
			dp.Timestamp = time.Now()
		}
	}()

	fmt3, err := parseFormat3(mfID, data)
	if err == nil {
		return fmt3, nil
	}
	errs = append(errs, fmt.Sprintf("parseFormat3 failed: %v", err))

	fmt5, err := parseFormat5(mfID, data)
	if err == nil {
		return fmt5, nil
	}
	errs = append(errs, fmt.Sprintf("parseFormat5 failed: %v", err))

	return nil, errors.New(strings.Join(errs, "\n"))
}
