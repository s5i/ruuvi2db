package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/s5i/ruuvi2db/data"
)

type format3 struct {
	DataFormat uint8
	Humid      uint8
	Temp       uint8
	TempFrac   uint8
	Pres       uint16
	AccX       int16
	AccY       int16
	AccZ       int16
	Batt       uint16
}

func parseFormat3(mfID uint16, raw []byte) (*data.Point, error) {
	if gotMFID, wantMFID := mfID, uint16(0x0499); gotMFID != wantMFID {
		return nil, fmt.Errorf("mfID mismatch (got %X, want %X)", gotMFID, wantMFID)
	}
	if gotLen, wantLen := len(raw), 14; gotLen < wantLen {
		return nil, fmt.Errorf("packet length mismatch (got %d, want at least %d)", gotLen, wantLen)
	}
	if gotID, wantID := int(raw[0]), 3; gotID != wantID {
		return nil, fmt.Errorf("format mismatch (got %d, want %d)", gotID, wantID)
	}

	var packet format3
	if err := binary.Read(bytes.NewReader(raw[0:14]), binary.BigEndian, &packet); err != nil {
		return nil, fmt.Errorf("binary.Read failed: %v", err)
	}

	temp := func(t uint8, f uint8) float64 {
		sign := float64(1)
		if (t & (1 << 7)) > 0 {
			sign = float64(-1)
		}
		t &^= 1 << 7

		return (float64(t) + float64(f)/100.0) * sign
	}

	return &data.Point{
		Temperature: temp(packet.Temp, packet.TempFrac),
		Humidity:    float64(packet.Humid) / 2.0,
		Pressure:    (float64(packet.Pres) + 50000.0) / 100.0,
		Battery:     float64(packet.Batt),
	}, nil
}
