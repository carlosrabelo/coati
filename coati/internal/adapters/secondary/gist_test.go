package secondary

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGistFetcher_Fetch(t *testing.T) {
	t.Run("named file", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/gists/abc", r.URL.Path)
			assert.Equal(t, "token secret", r.Header.Get("Authorization"))
			assert.Equal(t, "application/vnd.github.v3+json", r.Header.Get("Accept"))
			_ = json.NewEncoder(w).Encode(map[string]any{
				"files": map[string]any{
					"hosts.yaml": map[string]string{"content": "hosts: []"},
					"other.yaml": map[string]string{"content": "other"},
				},
			})
		}))
		t.Cleanup(srv.Close)

		f := NewGistFetcher()
		f.apiBase = srv.URL
		f.httpClient = srv.Client()

		got, err := f.Fetch("abc", "secret", "hosts.yaml")
		require.NoError(t, err)
		assert.Equal(t, []byte("hosts: []"), got)
	})

	t.Run("first file alphabetically when gistFile empty", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"files": map[string]any{
					"z.yaml": map[string]string{"content": "z"},
					"a.yaml": map[string]string{"content": "a-content"},
				},
			})
		}))
		t.Cleanup(srv.Close)

		f := NewGistFetcher()
		f.apiBase = srv.URL
		f.httpClient = srv.Client()

		got, err := f.Fetch("abc", "secret", "")
		require.NoError(t, err)
		assert.Equal(t, []byte("a-content"), got)
	})

	t.Run("status error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("missing"))
		}))
		t.Cleanup(srv.Close)

		f := NewGistFetcher()
		f.apiBase = srv.URL
		f.httpClient = srv.Client()

		_, err := f.Fetch("abc", "secret", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "github api error (status: 404): missing")
	})

	t.Run("missing named file", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"files": map[string]any{
					"only.yaml": map[string]string{"content": "x"},
				},
			})
		}))
		t.Cleanup(srv.Close)

		f := NewGistFetcher()
		f.apiBase = srv.URL
		f.httpClient = srv.Client()

		_, err := f.Fetch("abc", "secret", "missing.yaml")
		require.Error(t, err)
		assert.Equal(t, `file "missing.yaml" not found in gist abc`, err.Error())
	})

	t.Run("empty gist", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"files": map[string]any{},
			})
		}))
		t.Cleanup(srv.Close)

		f := NewGistFetcher()
		f.apiBase = srv.URL
		f.httpClient = srv.Client()

		_, err := f.Fetch("abc", "secret", "")
		require.Error(t, err)
		assert.Equal(t, "gist contains no files", err.Error())
	})
}

func TestGistPusher_Push(t *testing.T) {
	t.Run("named file patch", func(t *testing.T) {
		var gotBody map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPatch, r.Method)
			assert.Equal(t, "/gists/abc", r.URL.Path)
			assert.Equal(t, "token secret", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			raw, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(raw, &gotBody))
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(srv.Close)

		p := NewGistPusher()
		p.apiBase = srv.URL
		p.httpClient = srv.Client()

		err := p.Push("abc", "secret", "hosts.yaml", []byte("content"))
		require.NoError(t, err)
		files := gotBody["files"].(map[string]any)
		entry := files["hosts.yaml"].(map[string]any)
		assert.Equal(t, "content", entry["content"])
	})

	t.Run("empty gistFile resolves first alphabetically", func(t *testing.T) {
		var methods []string
		var patchBody map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			methods = append(methods, r.Method)
			switch r.Method {
			case http.MethodGet:
				_ = json.NewEncoder(w).Encode(map[string]any{
					"files": map[string]any{
						"z.yaml": map[string]string{"content": "z"},
						"a.yaml": map[string]string{"content": "a"},
					},
				})
			case http.MethodPatch:
				raw, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.NoError(t, json.Unmarshal(raw, &patchBody))
				w.WriteHeader(http.StatusOK)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		}))
		t.Cleanup(srv.Close)

		p := NewGistPusher()
		p.apiBase = srv.URL
		p.httpClient = srv.Client()

		err := p.Push("abc", "secret", "", []byte("new"))
		require.NoError(t, err)
		assert.Equal(t, []string{http.MethodGet, http.MethodPatch}, methods)
		files := patchBody["files"].(map[string]any)
		_, ok := files["a.yaml"]
		assert.True(t, ok)
	})

	t.Run("patch status error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte("nope"))
		}))
		t.Cleanup(srv.Close)

		p := NewGistPusher()
		p.apiBase = srv.URL
		p.httpClient = srv.Client()

		err := p.Push("abc", "secret", "hosts.yaml", []byte("x"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "github api error (status: 403): nope")
	})

	t.Run("empty gist when resolving filename", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"files": map[string]any{},
			})
		}))
		t.Cleanup(srv.Close)

		p := NewGistPusher()
		p.apiBase = srv.URL
		p.httpClient = srv.Client()

		err := p.Push("abc", "secret", "", []byte("x"))
		require.Error(t, err)
		assert.Equal(t, "gist contains no files", err.Error())
	})
}
