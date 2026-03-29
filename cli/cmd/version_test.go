package cmd

import (
	"os"
	"strings"
	"testing"
)

func TestRunVersion(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := RunVersion([]string{})

	w.Close()
	os.Stdout = old

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	output := string(buf[:n])

	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	if !strings.Contains(output, "anyvali version") {
		t.Errorf("expected version output, got: %s", output)
	}

	if !strings.Contains(output, Version) {
		t.Errorf("expected version %s in output, got: %s", Version, output)
	}
}

func TestVersionConstant(t *testing.T) {
	if Version == "" {
		t.Error("Version constant should not be empty")
	}
}
