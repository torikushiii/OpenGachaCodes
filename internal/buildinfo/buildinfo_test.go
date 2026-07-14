package buildinfo

import "testing"

func TestUserAgent(t *testing.T) {
	original := CommitHash
	t.Cleanup(func() { CommitHash = original })

	CommitHash = ""
	if got, want := UserAgent(), "OpenGachaCodes (+https://github.com/torikushiii/OpenGachaCodes)"; got != want {
		t.Fatalf("UserAgent() = %q, want %q", got, want)
	}

	CommitHash = "abc123"
	if got, want := UserAgent(), "OpenGachaCodes@abc123 (+https://github.com/torikushiii/OpenGachaCodes)"; got != want {
		t.Fatalf("UserAgent() = %q, want %q", got, want)
	}
}
