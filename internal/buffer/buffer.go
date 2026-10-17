// Package buffer implements a mutex-free FIFO string buffer using channels.
package buffer

type Buffer struct {
	ch chan string
}

// constructor
func New(capacity int) *Buffer {
	return &Buffer{ch: make(chan string, capacity)}
}

func (b *Buffer) Push(s string) bool {
	select {
	case b.ch <- s:
		return true
	default:
		return false
	}
}

func (b *Buffer) Pull() (string, bool) {
	select {
	case s := <-b.ch:
		return s, true
	default:
		return "", false
	}
}

func (b *Buffer) Size() int {
	return len(b.ch)
}
