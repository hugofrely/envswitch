package spinner

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Spinner represents a CLI spinner
type Spinner struct {
	frames  []string
	message string
	stop    chan bool
	mu      sync.Mutex
	writer  io.Writer
	active  bool
}

// New creates a new spinner with default frames
func New(message string) *Spinner {
	return &Spinner{
		frames:  []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		message: message,
		stop:    make(chan bool),
		writer:  os.Stdout,
		active:  false,
	}
}

// Start begins the spinner animation
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()

	go func() {
		i := 0
		for {
			select {
			case <-s.stop:
				return
			default:
				s.mu.Lock()
				frame := s.frames[i%len(s.frames)]
				fmt.Fprintf(s.writer, "\r%s %s", frame, s.message)
				s.mu.Unlock()
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()
}

// Update changes the spinner message while it's running
func (s *Spinner) Update(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()
}

// Success stops the spinner and displays a success message
func (s *Spinner) Success(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return
	}

	s.active = false
	s.stop <- true
	fmt.Fprintf(s.writer, "\r✓ %s\n", message)
}

// Error stops the spinner and displays an error message
func (s *Spinner) Error(message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return
	}

	s.active = false
	s.stop <- true
	fmt.Fprintf(s.writer, "\r✗ %s\n", message)
}

// Stop stops the spinner without displaying a message
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return
	}

	s.active = false
	s.stop <- true
	fmt.Fprintf(s.writer, "\r")
}
