package update

import "testing"

func TestNewer(t *testing.T) {
	tests := []struct {
		latest, current string
		want            bool
	}{
		{"0.3.0", "0.2.0", true},
		{"v1.0.0", "0.9.9", true},
		{"1.2.0", "1.2.0", false},
		{"1.1.9", "1.2.0", false},
		{"invalid", "1.0.0", false},
	}
	for _, tt := range tests {
		if got := newer(tt.latest, tt.current); got != tt.want {
			t.Errorf("newer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
		}
	}
}

func TestReleaseVersion(t *testing.T) {
	if !isReleaseVersion("v1.2.3") {
		t.Fatal("expected tagged version to be recognized")
	}
	if isReleaseVersion("dev") || isReleaseVersion("abc123") {
		t.Fatal("development builds must not show update notices")
	}
}
