package shim

import "testing"

func TestIsShimInvocation(t *testing.T) {
	tests := []struct {
		argv0 string
		want  bool
	}{
		{"godot", true},
		{"/usr/local/bin/godot", true},
		{"/home/user/.gdt/shims/godot", true},
		{"gdt", false},
		{"/usr/local/bin/gdt", false},
		{"godot.exe", true},
		{"gdt.exe", false},
	}
	for _, tt := range tests {
		t.Run(tt.argv0, func(t *testing.T) {
			got := IsShimInvocation(tt.argv0)
			if got != tt.want {
				t.Errorf("IsShimInvocation(%q) = %v, want %v", tt.argv0, got, tt.want)
			}
		})
	}
}
