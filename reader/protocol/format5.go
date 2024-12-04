package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/s5i/ruuvi2db/data"
)

type format5 struct {
	DataFormat uint8
	Temp       int16
	Humid      uint16
	Pres       uint16
	AccX       int16
	AccY       int16
	AccZ       int16
	Batt       uint16
	MvCount    uint8
	Seq        uint16
	MAC        [6]uint8
}

// https://github.com/ruuvi/docs/blob/master/communication/bluetooth-advertisements/data-format-5-rawv2.md
func parseFormat5(mfID uint16, raw []byte) (*data.Point, error) {
	if gotMFID, wantMFID := mfID, uint16(0x0499); gotMFID != wantMFID {
		return nil, fmt.Errorf("mfID mismatch (got %X, want %X)", gotMFID, wantMFID)
	}
	if gotLen, wantLen := len(raw), 24; gotLen < wantLen {
		return nil, fmt.Errorf("packet length mismatch (got %d, want at least %d)", gotLen, wantLen)
	}
	if gotFormat, wantFormat := int(raw[0]), 5; gotFormat != wantFormat {
		return nil, fmt.Errorf("format mismatch (got %d, want %d)", gotFormat, wantFormat)
	}

	var packet format5
	if err := binary.Read(bytes.NewReader(raw[0:24]), binary.BigEndian, &packet); err != nil {
		return nil, fmt.Errorf("binary.Read failed: %v", err)
	}

	return &data.Point{
		Temperature: float64(packet.Temp) * 0.005,
		Humidity:    float64(packet.Humid) * 0.0025,
		Pressure:    (float64(packet.Pres) + 50000.0) / 100.0,
		Battery:     float64((packet.Batt&0xFFE0)>>5) + 1600.0,
	}, nil
}
