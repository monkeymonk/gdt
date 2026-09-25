package download

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

func VerifyChecksum(path string, expected string) error {
	return verifyHash(path, expected, sha512.New)
}

func VerifySHA256(path string, expected string) error {
	return verifyHash(path, expected, sha256.New)
}

func verifyHash(path string, expected string, newHash func() hash.Hash) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	h := newHash()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("checksum mismatch\n  expected: %s\n  actual:   %s", expected, actual)
	}
	return nil
}
