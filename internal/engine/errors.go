package engine

import "errors"

var (
	ErrAlreadyInstalled = errors.New("version already installed")
	ErrNoVersion        = errors.New("no godot version found; install one with: gdt install <version>")
	ErrChecksumMismatch = errors.New("checksum verification failed")
	ErrDownloadFailed   = errors.New("download failed")
)

type ActionableError struct {
	Err        error
	Suggestion string
}

func (e *ActionableError) Error() string { return e.Err.Error() }

// Actionable is the required constructor for ActionableError per RULES.md
// Conventions. It wraps an error with a concrete, actionable next step
// (suggestion) for the user. A raw &ActionableError{} literal is against
// convention — always use Actionable() to construct. Use this whenever there
// is a non-obvious recovery path beyond the error text itself.
func Actionable(err error, suggestion string) *ActionableError {
	return &ActionableError{Err: err, Suggestion: suggestion}
}
func (e *ActionableError) Unwrap() error { return e.Err }
