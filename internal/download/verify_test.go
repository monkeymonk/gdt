package download

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifySHA256(t *testing.T) {
	content := []byte("known deterministic content for sha256 verification")

	dir := t.TempDir()
	path := filepath.Join(dir, "artifact")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}

	sum := sha256.Sum256(content)
	correctHex := hex.EncodeToString(sum[:])

	if err := VerifySHA256(path, correctHex); err != nil {
		t.Errorf("expected nil error for correct checksum, got %v", err)
	}

	wrongHex := strings.Repeat("0", 64)
	if err := VerifySHA256(path, wrongHex); err == nil {
		t.Error("expected error for incorrect checksum, got nil")
	}
}
