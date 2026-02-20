package util

import (
	"testing"
)

func TestFromCwd(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/home/user/my-app", "my-app"},
		{"/home/user/My App", "my-app"},
		{"/home/user/123abc", "123abc"},
		{"/home/user/---", "project"},
		{"", "project"},
		{"/home/user/Hello World 123", "hello-world-123"},
		{"/home/user/a..b..c", "a-b-c"},
		{"/home/user/UPPERCASE", "uppercase"},
		{"/home/user/special!@#chars", "special-chars"},
		{"/home/user/a", "a"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := FromCwd(tt.input)
			if got != tt.want {
				t.Errorf("FromCwd(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDeconflict(t *testing.T) {
	taken := map[string]bool{
		"api":   true,
		"api-2": true,
	}

	got := Deconflict("api", taken)
	if got != "api-3" {
		t.Errorf("Deconflict(api) = %q, want api-3", got)
	}

	got = Deconflict("web", taken)
	if got != "web" {
		t.Errorf("Deconflict(web) = %q, want web", got)
	}
}
