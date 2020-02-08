package server

import (
	"bytes"
	"io"
	"net"
	"testing"
	"time"

	"github.com/spin-org/thermomatic/internal/client"
)

// This test uses the below fakeConn. This easily isolates the test. However,
// this doesn't provide coverage for the Accept loop in Serve() (in this case,
// it's fairly simple). For more realism and coverage, you could actually make
// the server listen on a port available for testing and do a very similar
// test. The benchmark at the bottom uses that method instead.

type fakeConn struct {
	*bytes.Reader
}

func (*fakeConn) Write([]byte) (int, error) {
	return 0, io.EOF
}

func (*fakeConn) RemoteAddr() net.Addr {
	return nil
}
func (*fakeConn) LocalAddr() net.Addr {
	return nil
}
func (*fakeConn) SetDeadline(time.Time) error {
	return nil
}
func (*fakeConn) SetReadDeadline(time.Time) error {
	return nil
}
func (*fakeConn) SetWriteDeadline(time.Time) error {
	return nil
}

func (*fakeConn) Close() error {
	return nil
}

func TestServerServe(t *testing.T) {
	var output bytes.Buffer
	expectedOutput := []byte("1257894000000000000,490154203237518,67.77,2.63555,33.41,44.4,0.2566\n")
	timestamp := time.Unix(1257894000, 0)
	reading := &client.Reading{
		Temperature:  67.77,
		Altitude:     2.63555,
		Latitude:     33.41,
		Longitude:    44.4,
		BatteryLevel: 0.2566,
	}
	s := &Server{
		lis:    nil,
		output: &output,
		timeNow: func() time.Time {
			return timestamp
		},
	}
	var input bytes.Buffer
	input.Write([]byte("490154203237518"))
	input.Write(reading.EncodeForTest())
	conn := &fakeConn{
		bytes.NewReader(input.Bytes()),
	}
	s.handleConn(conn)
	if !bytes.Equal(output.Bytes(), expectedOutput) {
		t.Errorf("Wrong output from handleConn. got %q want %q", output.Bytes(), expectedOutput)
	}
}

type nullWriter struct{}

func (w nullWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func BenchmarkServer(b *testing.B) {
	// This benchmark shows how fast the server can handle readings from a
	// single source. This has limited usefulness since in the real world,
	// readings would trickle in from many sources at the same time.

	// I'm running out of time so I'll do this instead since it's simpler. If
	// you wanted to test how many clients a server can handle, you might not
	// want to use testing.B since it needs elapsed time to be correlated with
	// b.N or else it won't converge.
	lis, err := net.Listen("tcp", "[::1]:0")
	if err != nil {
		b.Fatalf("Error listening: %v", err)
	}
	closed := false
	defer func() {
		closed = true
		lis.Close()
	}()
	s := New(lis)
	go func() {
		if err := s.Serve(); err != nil && !closed {
			b.Fatalf("Error in Serve: %v", err)
		}
	}()
	s.output = nullWriter{}

	r := client.Reading{
		BatteryLevel: 1,
	}
	readingBuf := r.EncodeForTest()
	loginBuf := []byte("490154203237518")

	conn, err := net.Dial("tcp", lis.Addr().String())
	if err != nil {
		b.Fatalf("Error in Dial: %v", err)
	}
	if _, err := conn.Write(loginBuf); err != nil {
		b.Fatalf("Error writing login: %v", err)
	}

	b.SetBytes(int64(len(readingBuf)))
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, err := conn.Write(readingBuf); err != nil {
			b.Fatalf("Error writing reading: %v", err)
		}
	}
}
