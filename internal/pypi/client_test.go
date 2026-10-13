package pypi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pypi/demo-pkg/json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(sampleJSON))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, HTTP: srv.Client()}
	ps, err := c.FetchProject(context.Background(), "demo-pkg")
	if err != nil {
		t.Fatalf("FetchProject: %v", err)
	}
	if ps.TotalSize != 4500 {
		t.Errorf("TotalSize = %d, want 4500", ps.TotalSize)
	}
}

func TestFetchProjectNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, HTTP: srv.Client()}
	if _, err := c.FetchProject(context.Background(), "missing"); err == nil {
		t.Fatal("expected error for 404, got nil")
	}
}

func TestDiscoverUserProjects(t *testing.T) {
	page := `<html><body>
	  <a href="/project/gllm-inference-binary/">gllm-inference-binary</a>
	  <a href="/project/gllm-core-binary/">gllm-core-binary</a>
	  <a href="/project/gllm-inference-binary/">dup should dedupe</a>
	  <a href="/help/">not a project</a>
	</body></html>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/glx-sre/" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(page))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, HTTP: srv.Client()}
	got, err := c.DiscoverUserProjects(context.Background(), "glx-sre")
	if err != nil {
		t.Fatalf("DiscoverUserProjects: %v", err)
	}
	want := []string{"gllm-core-binary", "gllm-inference-binary"} // sorted, deduped
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
