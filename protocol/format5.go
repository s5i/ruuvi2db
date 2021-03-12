package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/s5i/ruuvi2db/data"
)

type format5 struct {
	MfID       uint16
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

func parseFormat5(raw []byte) (*data.Point, error) {
	if gotLen, wantLen := len(raw), 26; gotLen < wantLen {
		return nil, fmt.Errorf("packet length mismatch (got %d, want at least %d)", gotLen, wantLen)
	}
	if gotMSD, wantMSD := binary.LittleEndian.Uint16(raw[0:2]), uint16(0x0499); gotMSD != wantMSD {
		return nil, fmt.Errorf("MSD mismatch (got %X, want %X)", gotMSD, wantMSD)
	}
	if gotID, wantID := int(raw[2]), 5; gotID != wantID {
		return nil, fmt.Errorf("Manufacturer ID mismatch (got %d, want %d)", gotID, wantID)
	}

	var packet format5
	if err := binary.Read(bytes.NewReader(raw[0:26]), binary.BigEndian, &packet); err != nil {
		return nil, fmt.Errorf("binary.Read failed: %v", err)
	}

	return &data.Point{
		Temperature: float64(packet.Temp) * 0.005,
		Humidity:    float64(packet.Humid) * 0.0025,
		Pressure:    (float64(packet.Pres) + 50000.0) / 100.0,
		Battery:     float64((packet.Batt&0xFFE0)>>5) + 1600.0,
	}, nil
}
