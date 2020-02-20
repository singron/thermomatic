package server

import (
	"bytes"
	"log"
	"net"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spin-org/thermomatic/internal/client"
)

func TestServerServe(t *testing.T) {
	var output bytes.Buffer
	expectedOutput := []byte("1257894000000000000,490154203237518,67.77,2.63555,33.41,44.4,0.2566\n")
	timestamp := time.Unix(1257894000, 0)

	withServer(t, func(addr string, s *Server) {
		s.output = &output
		s.timeNow = func() time.Time {
			return timestamp
		}
		reading := &client.Reading{
			Temperature:  67.77,
			Altitude:     2.63555,
			Latitude:     33.41,
			Longitude:    44.4,
			BatteryLevel: 0.2566,
		}
		var input bytes.Buffer
		input.Write([]byte("490154203237518"))
		input.Write(reading.EncodeForTest())
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatalf("Error dialing server: %v", err)
		}
		defer conn.Close()
		if _, err := conn.Write(input.Bytes()); err != nil {
			t.Fatalf("Error writing input: %v", err)
		}
	})
	if !bytes.Equal(output.Bytes(), expectedOutput) {
		t.Errorf("Wrong output from handleConn. got %q want %q", output.Bytes(), expectedOutput)
	}
}

type nullWriter struct{}

func (w nullWriter) Write(b []byte) (int, error) {
	return len(b), nil
}

func withServer(t testing.TB, f func(addr string, s *Server)) {
	addr, err := net.ResolveTCPAddr("tcp", "[::1]:0")
	if err != nil {
		t.Fatal(err)
	}
	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatalf("Error listening: %v", err)
	}
	closed := int32(0)
	s := New(lis)
	defer s.wg.Wait()
	serveC := make(chan bool)
	defer func() {
		<-serveC
	}()
	defer func() {
		atomic.StoreInt32(&closed, 1)
		lis.Close()
	}()
	go func() {
		defer close(serveC)
		err := s.Serve()
		if err != nil && atomic.LoadInt32(&closed) != 1 {
			t.Fatalf("Error in Serve: %v", err)
		}
	}()
	f(lis.Addr().String(), s)
}

func BenchmarkServer(b *testing.B) {
	// This benchmark shows how fast the server can handle readings from a
	// single source. This has limited usefulness since in the real world,
	// readings would trickle in from many sources at the same time.

	// I'm running out of time so I'll do this instead since it's simpler. If
	// you wanted to test how many clients a server can handle, you might not
	// want to use testing.B since it needs elapsed time to be correlated with
	// b.N or else it won't converge.
	withServer(b, func(addr string, s *Server) {
		s.output = nullWriter{}
		r := client.Reading{
			BatteryLevel: 1,
		}
		readingBuf := r.EncodeForTest()
		loginBuf := []byte("490154203237518")

		conn, err := net.Dial("tcp", addr)
		if err != nil {
			b.Fatalf("Error in Dial: %v", err)
		}
		defer conn.Close()
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
	})

}

func BenchmarkServerLogin(b *testing.B) {
	withServer(b, func(addr string, s *Server) {
		s.output = nullWriter{}
		r := client.Reading{
			BatteryLevel: 1,
		}
		readingBuf := r.EncodeForTest()
		loginBuf := []byte("490154203237518")

		log.SetOutput(nullWriter{})
		defer log.SetOutput(os.Stderr)

		b.SetBytes(int64(len(readingBuf) + len(loginBuf)))
		b.ReportAllocs()
		b.ResetTimer()

		b.SetParallelism(10)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				conn, err := net.Dial("tcp", addr)
				if err != nil {
					b.Fatalf("Error in Dial: %v", err)
				}
				if _, err := conn.Write(loginBuf); err != nil {
					b.Fatalf("Error writing login: %v", err)
				}
				if _, err := conn.Write(readingBuf); err != nil {
					b.Fatalf("Error writing reading: %v", err)
				}
				conn.Close()
			}
		})
	})

}
