package client

import (
	"encoding/binary"
	"math"
)

// Reading is the set of device readings.
type Reading struct {
	// Temperature denotes the temperature reading of the message.
	Temperature float64

	// Altitude denotes the altitude reading of the message.
	Altitude float64

	// Latitude denotes the latitude reading of the message.
	Latitude float64

	// Longitude denotes the longitude reading of the message.
	Longitude float64

	// BatteryLevel denotes the battery level reading of the message.
	BatteryLevel float64
}

func readBEFloat64(b []byte) float64 {
	u := binary.BigEndian.Uint64(b)
	return math.Float64frombits(u)
}

// Decode decodes the reading message payload in the given b into r.
//
// If any of the fields are outside their valid min/max ranges ok will be unset.
//
// Decode does NOT allocate under any condition. Additionally, it panics if b
// isn't at least 40 bytes long.
func (r *Reading) Decode(b []byte) (ok bool) {
	if len(b) < 40 {
		panic("Reading buffer is too short")
	}
	ok = true
	r.Temperature = readBEFloat64(b[0:8])
	if math.IsNaN(r.Temperature) || r.Temperature < -300 || r.Temperature > 300 {
		ok = false
	}
	r.Altitude = readBEFloat64(b[8:16])
	if math.IsNaN(r.Altitude) || r.Altitude < -20000 || r.Altitude > 20000 {
		ok = false
	}
	r.Latitude = readBEFloat64(b[16:24])
	if math.IsNaN(r.Latitude) || r.Latitude < -90 || r.Latitude > 90 {
		ok = false
	}
	r.Longitude = readBEFloat64(b[24:32])
	if math.IsNaN(r.Longitude) || r.Longitude < -180 || r.Longitude > 180 {
		ok = false
	}
	r.BatteryLevel = readBEFloat64(b[32:40])
	// This case is a little different since battery level can't be 0
	if math.IsNaN(r.BatteryLevel) || r.BatteryLevel <= 0 || r.BatteryLevel > 100 {
		ok = false
	}

	return ok
}

// EncodeForTest encodes Reading so that it can be decoded with Decode. It
// makes no attempt to avoid allocations, but is still useful for tests.
func (r *Reading) EncodeForTest() []byte {
	enc := binary.BigEndian
	b := make([]byte, 40)
	enc.PutUint64(b[0:8], math.Float64bits(r.Temperature))
	enc.PutUint64(b[8:16], math.Float64bits(r.Altitude))
	enc.PutUint64(b[16:24], math.Float64bits(r.Latitude))
	enc.PutUint64(b[24:32], math.Float64bits(r.Longitude))
	enc.PutUint64(b[32:40], math.Float64bits(r.BatteryLevel))
	return b
}
