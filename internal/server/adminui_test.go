package server_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/han/go-ecommerce/internal/server"
)

func TestAdminUIAssetsAndAPICatalog(t *testing.T) {
	f := setup(t)
	r := server.NewRouter(f.db, f.jwt)
	for path, kind := range map[string]string{"/admin/": "text/html", "/admin/app.js": "text/javascript", "/admin/style.css": "text/css", "/admin/api.json": "application/json"} {
		res := request(t, r, "", http.MethodGet, path, nil)
		if res.Code != 200 || !strings.Contains(res.Header().Get("Content-Type"), kind) {
			t.Fatalf("%s: %d %s", path, res.Code, res.Header())
		}
		if res.Header().Get("Content-Security-Policy") == "" {
			t.Fatalf("%s missing CSP", path)
		}
	}
	if res := request(t, r, "", http.MethodGet, "/admin", nil); res.Code != 307 || res.Header().Get("Location") != "/admin/" {
		t.Fatal("missing admin redirect")
	}
	if res := request(t, r, "", http.MethodGet, "/admin/../../.env", nil); res.Code != 404 {
		t.Fatal("unexpected asset exposure")
	}
	catalog := request(t, r, "", http.MethodGet, "/admin/api.json", nil)
	var entries []struct {
		Method string `json:"method"`
		Path   string `json:"path"`
		Note   string `json:"note"`
	}
	requireNoError(t, json.Unmarshal(catalog.Body.Bytes(), &entries))
	actual := map[string]bool{}
	for _, route := range r.Routes() {
		if route.Path != "/admin" && !strings.HasPrefix(route.Path, "/admin/") {
			actual[route.Method+" "+route.Path] = true
		}
	}
	for _, e := range entries {
		key := e.Method + " " + e.Path
		if !actual[key] || e.Note == "" {
			t.Fatalf("invalid/duplicate catalog entry: %s", key)
		}
		delete(actual, key)
	}
	if len(actual) != 0 {
		t.Fatalf("undocumented endpoints: %v", actual)
	}
	// Public HTML does not grant access to the underlying data.
	for _, path := range []string{"/api/v1/admin/questions", "/api/v1/admin/tests"} {
		if res := request(t, r, "", http.MethodGet, path, nil); res.Code != 401 {
			t.Fatalf("anonymous data access: %s %d", path, res.Code)
		}
		learner, err := f.jwt.GenerateAccessToken(f.user.ID)
		requireNoError(t, err)
		if res := request(t, r, learner, http.MethodGet, path, nil); res.Code != 403 {
			t.Fatalf("learner data access: %s %d", path, res.Code)
		}
	}
}
