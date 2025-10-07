package spinner

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	spin := New("test message")
	if spin == nil {
		t.Fatal("New() returned nil")
	}
	if spin.message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", spin.message)
	}
	if len(spin.frames) == 0 {
		t.Error("Spinner has no frames")
	}
}

func TestSpinnerStartStop(t *testing.T) {
	var buf bytes.Buffer
	spin := New("testing")
	spin.writer = &buf

	spin.Start()
	if !spin.active {
		t.Error("Spinner should be active after Start()")
	}

	// Let it spin a bit
	time.Sleep(200 * time.Millisecond)

	spin.Stop()
	time.Sleep(50 * time.Millisecond) // Wait for goroutine to finish

	if spin.active {
		t.Error("Spinner should not be active after Stop()")
	}

	output := buf.String()
	if !strings.Contains(output, "testing") {
		t.Errorf("Output should contain 'testing', got: %s", output)
	}
}

func TestSpinnerUpdate(t *testing.T) {
	var buf bytes.Buffer
	spin := New("initial")
	spin.writer = &buf

	spin.Start()
	time.Sleep(100 * time.Millisecond)

	spin.Update("updated")
	time.Sleep(100 * time.Millisecond)

	spin.Stop()
	time.Sleep(50 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "updated") {
		t.Errorf("Output should contain 'updated', got: %s", output)
	}
}

func TestSpinnerSuccess(t *testing.T) {
	var buf bytes.Buffer
	spin := New("working")
	spin.writer = &buf

	spin.Start()
	time.Sleep(100 * time.Millisecond)

	spin.Success("completed successfully")
	time.Sleep(50 * time.Millisecond)

	if spin.active {
		t.Error("Spinner should not be active after Success()")
	}

	output := buf.String()
	if !strings.Contains(output, "✓") {
		t.Errorf("Output should contain success checkmark, got: %s", output)
	}
	if !strings.Contains(output, "completed successfully") {
		t.Errorf("Output should contain success message, got: %s", output)
	}
}

func TestSpinnerError(t *testing.T) {
	var buf bytes.Buffer
	spin := New("working")
	spin.writer = &buf

	spin.Start()
	time.Sleep(100 * time.Millisecond)

	spin.Error("failed with error")
	time.Sleep(50 * time.Millisecond)

	if spin.active {
		t.Error("Spinner should not be active after Error()")
	}

	output := buf.String()
	if !strings.Contains(output, "✗") {
		t.Errorf("Output should contain error mark, got: %s", output)
	}
	if !strings.Contains(output, "failed with error") {
		t.Errorf("Output should contain error message, got: %s", output)
	}
}

func TestSpinnerMultipleStarts(t *testing.T) {
	var buf bytes.Buffer
	spin := New("test")
	spin.writer = &buf

	spin.Start()
	time.Sleep(50 * time.Millisecond)

	// Starting again should not cause issues
	spin.Start()
	time.Sleep(50 * time.Millisecond)

	spin.Stop()
	time.Sleep(50 * time.Millisecond)

	if spin.active {
		t.Error("Spinner should not be active after Stop()")
	}
}
