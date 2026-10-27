package readiness

import (
	"os"
	"testing"
)

func TestReadinessRequiresSuccessfulInitialization(t *testing.T) {
	dir := t.TempDir()
	if err := checkAt(dir, 42); err == nil {
		t.Fatal("missing readiness accepted")
	}
	if err := markAt(dir, 42); err != nil {
		t.Fatal(err)
	}
	if err := checkAt(dir, 42); err != nil {
		t.Fatal(err)
	}
	if err := checkAt(dir, 43); err == nil {
		t.Fatal("another process's readiness accepted")
	}
	if err := os.WriteFile(recordPath(dir, 42), []byte("43"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := checkAt(dir, 42); err == nil {
		t.Fatal("mismatched readiness accepted")
	}
	if err := markAt(dir, 0); err == nil {
		t.Fatal("invalid PID accepted")
	}
}
