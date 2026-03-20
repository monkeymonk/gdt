package engine

import "errors"

var (
	ErrAlreadyInstalled = errors.New("version already installed")
	ErrNotInstalled     = errors.New("version not installed")
	ErrNoVersion        = errors.New("no godot version found; install one with: gdt install <version>")
	ErrChecksumMismatch = errors.New("checksum verification failed")
	ErrDownloadFailed   = errors.New("download failed")
)

type ActionableError struct {
	Err        error
	Suggestion string
}

func (e *ActionableError) Error() string { return e.Err.Error() }
func (e *ActionableError) Unwrap() error { return e.Err }
