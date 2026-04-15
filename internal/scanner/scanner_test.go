package scanner

import (
	"runtime"
	"testing"
)

func TestNew_ReturnsCorrectScanner(t *testing.T) {
	scn := New()
	if scn == nil {
		t.Fatal("New() returned nil")
	}

	switch runtime.GOOS {
	case "linux":
		if _, ok := scn.(*LinuxScanner); !ok {
			t.Errorf("on linux, expected *LinuxScanner, got %T", scn)
		}
	case "darwin":
		if _, ok := scn.(*DarwinScanner); !ok {
			t.Errorf("on darwin, expected *DarwinScanner, got %T", scn)
		}
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.RangeStart != 1 {
		t.Errorf("RangeStart = %d, want 1", opts.RangeStart)
	}
	if opts.RangeEnd != 65535 {
		t.Errorf("RangeEnd = %d, want 65535", opts.RangeEnd)
	}
	if opts.Port != 0 {
		t.Errorf("Port = %d, want 0", opts.Port)
	}
	if opts.TCPOnly {
		t.Error("TCPOnly should be false by default")
	}
	if opts.UDPOnly {
		t.Error("UDPOnly should be false by default")
	}
	if opts.ListenOnly {
		t.Error("ListenOnly should be false by default")
	}
	if opts.Filter != "" {
		t.Errorf("Filter = %q, want empty", opts.Filter)
	}
}
