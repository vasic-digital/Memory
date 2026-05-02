package memfd

import (
	"sync"
)

// PSC is a Producer-Shared-Consumer ring buffer.
type PSC struct {
	mu       sync.Mutex
	buf      []byte
	cap      int
	readOff  int
	writeOff int
	closed   bool
}

// NewPSC creates a new PSC with the given capacity.
func NewPSC(capacity int) *PSC {
	return &PSC{
		buf: make([]byte, capacity),
		cap: capacity,
	}
}

// Write copies data into the ring buffer.
func (p *PSC) Write(data []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closed {
		return 0, nil
	}
	n := copy(p.buf[p.writeOff:], data)
	p.writeOff += n
	if p.writeOff >= p.cap {
		p.writeOff = 0
	}
	return n, nil
}

// Read copies data out of the ring buffer.
func (p *PSC) Read(out []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.readOff >= p.writeOff {
		return 0, nil
	}
	n := copy(out, p.buf[p.readOff:p.writeOff])
	p.readOff += n
	return n, nil
}

// Close marks the buffer as closed.
func (p *PSC) Close() error {
	p.mu.Lock()
	p.closed = true
	p.mu.Unlock()
	return nil
}
