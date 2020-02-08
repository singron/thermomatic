// Package imei implements an IMEI decoder.
package imei

// NOTE: for more information about IMEI codes and their structure you may
// consult with:
//
// https://en.wikipedia.org/wiki/International_Mobile_Equipment_Identity.

import (
	"errors"
)

var (
	ErrInvalid  = errors.New("imei: invalid ")
	ErrChecksum = errors.New("imei: invalid checksum")
)

// Decode returns the IMEI code contained in the first 15 bytes of b.
//
// In case b isn't strictly composed of digits, the returned error will be
// ErrInvalid.
//
// In case b's checksum is wrong, the returned error will be ErrChecksum.
//
// Decode does NOT allocate under any condition. Additionally, it panics if b
// isn't at least 15 bytes long.
func Decode(b []byte) (code uint64, err error) {
	if len(b) < 15 {
		panic("imei buffer too short")
	}
	// Don't use strconv.ParseUint since that takes a string that would need
	// to be allocated.
	var n uint64
	var checksum uint64 // Luhn checksum
	for i, c := range b[:15] {
		if c < '0' || c > '9' {
			return 0, ErrInvalid
		}
		b := uint64(c - '0')
		// A 14 digit base-10 number cannot overflow a uint64, so no need to
		// check.
		n *= 10
		n += b

		if i%2 == 1 {
			// Each second digit from the right is doubled. In a 15 digit
			// number, that's each odd digit.
			b += b
			if b >= 10 {
				b -= 9
			}
		}
		checksum += b
	}

	if checksum%10 != 0 {
		return 0, ErrChecksum
	}
	return n, nil
}
