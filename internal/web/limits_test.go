package web

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestInsideDir(t *testing.T) {
	root := t.TempDir()
	served := filepath.Join(root, "served")
	if err := os.Mkdir(served, 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(served, link); err != nil {
		t.Skip("symlinks unavailable")
	}
	for _, tc := range []struct {
		file string
		want bool
	}{
		{filepath.Join(served, "ca.key"), true},
		{filepath.Join(link, "ca.key"), true},
		{filepath.Join(root, "ca.key"), false},
		{filepath.Join(root, "served-other", "ca.key"), false},
	} {
		if tc.file == filepath.Join(root, "served-other", "ca.key") {
			if err := os.Mkdir(filepath.Dir(tc.file), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		got, err := insideDir(tc.file, served)
		if err != nil {
			t.Fatal(err)
		}
		if got != tc.want {
			t.Errorf("insideDir(%s) = %v, want %v", tc.file, got, tc.want)
		}
	}
}

func TestRequestedDays(t *testing.T) {
	for _, tc := range []struct {
		query string
		ok    bool
	}{
		{"", true},
		{"?days=1095", true},
		{"?days=1096", false},
		{"?days=0", false},
		{"?days=-5", false},
		{"?days=x", false},
	} {
		r := httptest.NewRequest("GET", "/crt"+tc.query, nil)
		if _, ok := requestedDays(httptest.NewRecorder(), r); ok != tc.ok {
			t.Errorf("%q: ok = %v, want %v", tc.query, ok, tc.ok)
		}
	}
}
