package metadata

import "testing"

func TestFindChecksum(t *testing.T) {
	content := `abc123def456  Godot_v4.3-stable_linux.x86_64.zip
789xyz000111  Godot_v4.3-stable_win64.exe.zip
deadbeef1234  Godot_v4.3-stable_macos.universal.zip`

	tests := []struct {
		name     string
		content  string
		artifact string
		want     string
	}{
		{"match found", content, "Godot_v4.3-stable_linux.x86_64.zip", "abc123def456"},
		{"match second", content, "Godot_v4.3-stable_win64.exe.zip", "789xyz000111"},
		{"no match", content, "nonexistent.zip", ""},
		{"empty content", "", "anything.zip", ""},
		{"empty artifact", content, "", ""},
		{"partial name no match", content, "linux.x86_64.zip", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindChecksum(tt.content, tt.artifact)
			if got != tt.want {
				t.Errorf("FindChecksum() = %q, want %q", got, tt.want)
			}
		})
	}
}
