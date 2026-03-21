package plugins

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestParseStatusLine(t *testing.T) {
	tests := []struct {
		line   string
		status string
		msg    string
		ok     bool
	}{
		{"OK dotnet-sdk installed", "OK", "dotnet-sdk installed", true},
		{"WARN nuget outdated", "WARN", "nuget outdated", true},
		{"FAIL mono not found", "FAIL", "mono not found", true},
		{"", "", "", false},
		{"INVALID", "", "", false},
		{"OK", "OK", "", true},
	}
	for _, tt := range tests {
		status, msg, ok := ParseStatusLine(tt.line)
		if ok != tt.ok {
			t.Errorf("ParseStatusLine(%q): ok = %v, want %v", tt.line, ok, tt.ok)
			continue
		}
		if status != tt.status {
			t.Errorf("ParseStatusLine(%q): status = %q, want %q", tt.line, status, tt.status)
		}
		if msg != tt.msg {
			t.Errorf("ParseStatusLine(%q): msg = %q, want %q", tt.line, msg, tt.msg)
		}
	}
}

func TestParseStatusLines(t *testing.T) {
	output := "OK check-one passed\nWARN check-two slow\nFAIL check-three missing\n"
	results := ParseStatusLines(output)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0].Status != "OK" {
		t.Errorf("result[0].Status = %q, want OK", results[0].Status)
	}
	if results[1].Status != "WARN" {
		t.Errorf("result[1].Status = %q, want WARN", results[1].Status)
	}
	if results[2].Status != "FAIL" {
		t.Errorf("result[2].Status = %q, want FAIL", results[2].Status)
	}
}

func TestRunPluginSubcommand_Success(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	dir := t.TempDir()
	binPath := filepath.Join(dir, "test-plugin")
	script := "#!/bin/sh\necho \"OK check passed\"\n"
	os.WriteFile(binPath, []byte(script), 0755)

	out, err := RunPluginSubcommand(binPath, dir, nil, 5*time.Second, "doctor", "check")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "OK check passed") {
		t.Errorf("output = %q, want to contain 'OK check passed'", out)
	}
}

func TestRunPluginSubcommand_Timeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	dir := t.TempDir()
	binPath := filepath.Join(dir, "slow-plugin")
	script := "#!/bin/sh\nsleep 30\n"
	os.WriteFile(binPath, []byte(script), 0755)

	_, err := RunPluginSubcommand(binPath, dir, nil, 100*time.Millisecond, "doctor", "check")
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestRunPluginSubcommand_NonZeroExit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows")
	}

	dir := t.TempDir()
	binPath := filepath.Join(dir, "fail-plugin")
	script := "#!/bin/sh\necho \"FAIL broken\"\nexit 1\n"
	os.WriteFile(binPath, []byte(script), 0755)

	out, err := RunPluginSubcommand(binPath, dir, nil, 5*time.Second, "doctor", "check")
	if err == nil {
		t.Fatal("expected error for non-zero exit")
	}
	if !strings.Contains(out, "FAIL broken") {
		t.Errorf("output = %q, want to contain 'FAIL broken'", out)
	}
}
