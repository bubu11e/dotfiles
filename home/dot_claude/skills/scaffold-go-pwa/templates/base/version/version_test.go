package version_test

import (
	"runtime"
	"testing"

	"__MODULE__/version"
)

func TestGetReportsGoVersion(t *testing.T) {
	if got := version.Get().GoVersion; got != runtime.Version() {
		t.Errorf("GoVersion = %q, want %q", got, runtime.Version())
	}
}

func TestShortCommit(t *testing.T) {
	cases := map[string]string{
		"":             "",
		"abc":          "abc",
		"0123456789ab": "0123456789ab",
		"0123456789abcdef0123456789abcdef01234567": "0123456789ab",
	}
	for in, want := range cases {
		if got := version.ShortCommit(in); got != want {
			t.Errorf("ShortCommit(%q) = %q, want %q", in, got, want)
		}
	}
}
