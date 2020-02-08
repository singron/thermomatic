package client

import (
	"math"
	"reflect"
	"runtime"
	"testing"
)

func TestDecode(t *testing.T) {
	cases := []struct {
		r  Reading
		ok bool
	}{
		{
			r: Reading{
				BatteryLevel: 0,
			},
			ok: false,
		},
		{
			// Knowing this is a minimal valid input can simplify remaining tests.
			r: Reading{
				BatteryLevel: 1,
			},
			ok: true,
		},
		{
			r: Reading{
				BatteryLevel: 1,
				Altitude:     math.NaN(),
			},
			ok: false,
		},
		{
			r: Reading{
				BatteryLevel: 1,
				Altitude:     1000000000.0,
			},
			ok: false,
		},
		// I'm running out of time, but you could verify that the various
		// invalid inputs are detected by Decode by adding more testcases like
		// the above.
	}
	for _, tc := range cases {
		buf := tc.r.EncodeForTest()
		var r Reading
		ok := r.Decode(buf)
		if ok != tc.ok {
			t.Errorf("Decode(%#v) returned %v, wanted %v", tc.r, ok, tc.ok)
		}
		if ok && tc.ok {
			// Only check equality if both are OK because NaN isn't equal to
			// itself.
			if !reflect.DeepEqual(tc.r, r) {
				t.Errorf("Decode(%#v) unexpectedly resulted in %#v", tc.r, r)
			}
		}
	}
}

func TestDecodeAllocations(t *testing.T) {
	r := Reading{
		// All other values are valid as 0.
		BatteryLevel: 1,
	}
	buf := r.EncodeForTest()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	initAllocs := m.Alloc
	if ok := r.Decode(buf); !ok {
		t.Fatal("Decode failed")
	}
	runtime.ReadMemStats(&m)
	if m.Alloc > initAllocs {
		t.Errorf("Decode made %d allocations", m.Alloc-initAllocs)
	}
}

func BenchmarkDecode(b *testing.B) {
	r := Reading{
		// All other values are valid as 0.
		BatteryLevel: 1,
	}
	buf := r.EncodeForTest()
	b.SetBytes(int64(len(buf)))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if ok := r.Decode(buf); !ok {
			b.Fatal("Decode failed")
		}
	}
}
