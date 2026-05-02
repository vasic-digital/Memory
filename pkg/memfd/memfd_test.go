package memfd

import (
	"bytes"
	"testing"
)

func TestPSCWriteRead(t *testing.T) {
	p := NewPSC(1024)
	data := []byte("hello")
	n, err := p.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Expected %d bytes written, got %d", len(data), n)
	}

	out := make([]byte, 10)
	n, err = p.Read(out)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != len(data) {
		t.Fatalf("Expected %d bytes read, got %d", len(data), n)
	}
	if !bytes.Equal(out[:n], data) {
		t.Errorf("Expected %s, got %s", data, out[:n])
	}
}
