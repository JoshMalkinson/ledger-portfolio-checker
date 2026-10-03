package main

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHealthAndStaticFrontend(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<h1>Portfolio checker</h1>"), 0600); err != nil {
		t.Fatal(err)
	}
	h := newApp(dir)
	for _, tc := range []struct {
		path string
		code int
	}{{"/healthz", 200}, {"/", 200}, {"/missing", 404}} {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
		if w.Code != tc.code {
			t.Fatalf("%s: %d", tc.path, w.Code)
		}
	}
}
