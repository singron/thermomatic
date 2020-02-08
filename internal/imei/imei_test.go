package imei

import (
	"testing"
	//"github.com/spin-org/thermomatic/internal/common"
	"runtime"
)

func TestDecode(t *testing.T) {
	cases := []struct {
		b    []byte
		code uint64
		err  error
	}{
		{
			// From wikipedia
			b:    []byte("352099001761481"),
			code: 352099001761481,
		},
		{
			// From README
			b:    []byte("490154203237518"),
			code: 490154203237518,
		},
		{
			b:    []byte("352099001761481-ignorethis"),
			code: 352099001761481,
		},
		{
			b:   []byte("352099001761480"),
			err: ErrChecksum,
		},
		{
			b:   []byte("35a099001761480"),
			err: ErrInvalid,
		},
		{
			b:   []byte("35209900176148a"),
			err: ErrInvalid,
		},
	}

	for _, tc := range cases {
		code, err := Decode(tc.b)
		if tc.code != code || tc.err != err {
			t.Errorf("Unexpected result from Decode(%q): wanted (%v, %v) got (%v, %v)",
				tc.b, tc.code, tc.err, code, err)
		}
	}
}

func TestDecodeAllocations(t *testing.T) {
	buf := []byte("352099001761481")
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	initAllocs := m.Alloc
	_, err := Decode(buf)
	if err != nil {
		t.Fatalf("Unexpected failure in Decode: %v", err)
	}
	runtime.ReadMemStats(&m)
	if m.Alloc > initAllocs {
		t.Errorf("Decode made %d allocations", m.Alloc-initAllocs)
	}
}

func checkPanics(f func()) (paniced bool) {
	defer func() {
		if r := recover(); r != nil {
			paniced = true
		}
	}()
	f()
	return false
}

func TestDecodePanics(t *testing.T) {
	if !checkPanics(func() { Decode([]byte("1234")) }) {
		t.Error("Expected short imei buffer to panic in Decode")
	}
}

func BenchmarkDecode(b *testing.B) {
	b.ReportAllocs()
	buf := []byte("352099001761481")
	b.SetBytes(int64(len(buf)))
	for i := 0; i < b.N; i++ {
		_, err := Decode(buf)
		if err != nil {
			b.Fatalf("Unexpected failure in Decode: %v", err)
		}
	}
}
