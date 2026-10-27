// Package readiness records completion of core and node initialization.
// It does not change the node configuration or require panel support.
package readiness

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func directory() string {
	if runtime.GOOS == "linux" {
		return "/run/V2bX"
	}
	return filepath.Join(os.TempDir(), "V2bX-readiness")
}
func recordPath(dir string, pid int) string { return filepath.Join(dir, strconv.Itoa(pid)+".ready") }
func Mark(pid int) error                    { return markAt(directory(), pid) }
func Clear(pid int)                         { _ = os.Remove(recordPath(directory(), pid)) }
func Check(pid int) error                   { return checkAt(directory(), pid) }

func markAt(dir string, pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid readiness PID")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".ready-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err := fmt.Fprintln(file, pid); err != nil {
		file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), recordPath(dir, pid))
}
func checkAt(dir string, pid int) error {
	if pid <= 0 {
		return fmt.Errorf("invalid readiness PID")
	}
	data, err := os.ReadFile(recordPath(dir, pid))
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(data)) != strconv.Itoa(pid) {
		return fmt.Errorf("readiness PID mismatch")
	}
	return nil
}
