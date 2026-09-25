package plugins

import "testing"

func TestCheckRequiresGdt(t *testing.T) {
	tests := []struct {
		name       string
		constraint string
		gdtVersion string
		wantErr    bool
	}{
		{
			name:       "satisfied exact version",
			constraint: ">=0.3.0",
			gdtVersion: "0.3.0",
			wantErr:    false,
		},
		{
			name:       "satisfied newer version",
			constraint: ">=0.3.0",
			gdtVersion: "0.4.0",
			wantErr:    false,
		},
		{
			name:       "unsatisfied - real-world scenario",
			constraint: ">=1.0",
			gdtVersion: "0.3.0",
			wantErr:    true,
		},
		{
			name:       "no constraint declared",
			constraint: "",
			gdtVersion: "0.3.0",
			wantErr:    false,
		},
		{
			name:       "dev build always passes regardless of constraint",
			constraint: ">=1.0",
			gdtVersion: "dev",
			wantErr:    false,
		},
		{
			name:       "dev build passes even against far future constraint",
			constraint: ">=99.0.0",
			gdtVersion: "dev",
			wantErr:    false,
		},
		{
			name:       "empty gdtVersion passes",
			constraint: ">=1.0",
			gdtVersion: "",
			wantErr:    false,
		},
		{
			name:       "malformed constraint - wrong operator does not block",
			constraint: "^1.0",
			gdtVersion: "0.3.0",
			wantErr:    false,
		},
		{
			name:       "malformed constraint - garbage version does not block or panic",
			constraint: ">=not-a-version",
			gdtVersion: "0.3.0",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkRequiresGdt("testplugin", tt.constraint, tt.gdtVersion)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkRequiresGdt(%q, %q) = %v, wantErr %v", tt.constraint, tt.gdtVersion, err, tt.wantErr)
			}
		})
	}
}
