package cli

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestClassifyKillError(t *testing.T) {
	otherErr := errors.New("permission denied")

	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "nil error is not an issue", err: nil, want: nil},
		{name: "ErrProcessDone is not an issue", err: os.ErrProcessDone, want: nil},
		{name: "wrapped ErrProcessDone is not an issue", err: fmt.Errorf("kill: %w", os.ErrProcessDone), want: nil},
		{name: "other error is reported unchanged", err: otherErr, want: otherErr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyKillError(tt.err)
			if !errors.Is(got, tt.want) && got != tt.want {
				t.Errorf("classifyKillError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
