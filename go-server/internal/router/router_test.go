package router

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"ai-for-code/go-server/internal/activity"
	"ai-for-code/go-server/internal/handlers"
	"ai-for-code/go-server/internal/storage"
	"ai-for-code/go-server/internal/validation"
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

func newTestRouter(t *testing.T) http.Handler {
	t.Helper()
	db, err := storage.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	h := handlers.New(storage.NewUserRepository(db), activity.NewTracker(activity.DefaultLimit))
	return New(h)
}

func do(t *testing.T, h http.Handler, method, path, body string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var out map[string]any
	if strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("%s %s: invalid JSON %q", method, path, rec.Body.String())
		}
	}
	return rec.Code, out
}

func TestAddAndGetUser(t *testing.T) {
	h := newTestRouter(t)
	name := "x'); DROP TABLE users;--"

	code, body := do(t, h, "POST", "/adduser", fmt.Sprintf(`{"name": %q}`, "  "+name+"  "))
	if code != http.StatusCreated || body["name"] != name {
		t.Fatalf("add user: got %d %v", code, body)
	}
	id := int64(body["id"].(float64))

	code, body = do(t, h, "GET", fmt.Sprintf("/user/%d", id), "")
	if code != http.StatusOK || body["name"] != name {
		t.Fatalf("get user: got %d %v", code, body)
	}
}

func TestGetUserNotFound(t *testing.T) {
	h := newTestRouter(t)
	for _, path := range []string{"/user/999", "/user/abc", "/user/1%20OR%201=1"} {
		if code, body := do(t, h, "GET", path, ""); code != http.StatusNotFound || body["error"] != "user not found" {
			t.Errorf("%s: got %d %v", path, code, body)
		}
	}
}

func TestAddUserValidation(t *testing.T) {
	h := newTestRouter(t)
	cases := []string{
		``,
		`junk`,
		`null`,
		`[]`,
		`"name"`,
		`{}`,
		`{"name": null}`,
		`{"name": ""}`,
		`{"name": "   "}`,
		`{"name": 5}`,
		fmt.Sprintf(`{"name": %q}`, strings.Repeat("я", validation.MaxNameLength+1)),
	}
	for _, c := range cases {
		if code, body := do(t, h, "POST", "/adduser", c); code != http.StatusBadRequest || body["error"] == nil {
			t.Errorf("body %q: got %d %v", c, code, body)
		}
	}
}

func TestActiveUsers(t *testing.T) {
	h := newTestRouter(t)

	if code, body := do(t, h, "GET", "/active", ""); code != http.StatusOK || len(body["active"].([]any)) != 0 {
		t.Fatalf("empty active: got %d %v", code, body)
	}
	if code, _ := do(t, h, "POST", "/user/999/active", ""); code != http.StatusNotFound {
		t.Fatalf("unknown user: got %d", code)
	}

	const n = 20
	for i := range n {
		do(t, h, "POST", "/adduser", fmt.Sprintf(`{"name": "u%d"}`, i))
	}

	var wg sync.WaitGroup
	for id := 1; id <= n; id++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("POST", fmt.Sprintf("/user/%d/active", id), nil))
		}()
	}
	wg.Wait()

	code, body := do(t, h, "GET", "/active", "")
	if code != http.StatusOK || len(body["active"].([]any)) != activity.DefaultLimit {
		t.Fatalf("active after %d marks: got %d %v", n, code, body)
	}

	_, body = do(t, h, "POST", "/user/1/active", "")
	active := body["active"].([]any)
	if len(active) != activity.DefaultLimit || active[len(active)-1] != float64(1) {
		t.Fatalf("latest mark must be last: %v", active)
	}
}

func TestDocs(t *testing.T) {
	h := newTestRouter(t)
	cases := []struct {
		path, contentType, contains string
	}{
		{"/docs", "text/html", "SwaggerUIBundle"},
		{"/openapi.yaml", "application/yaml", "openapi: 3.1.0"},
	}
	for _, c := range cases {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", c.path, nil))
		if rec.Code != http.StatusOK ||
			!strings.HasPrefix(rec.Header().Get("Content-Type"), c.contentType) ||
			!strings.Contains(rec.Body.String(), c.contains) {
			t.Errorf("%s: got %d %q", c.path, rec.Code, rec.Header().Get("Content-Type"))
		}
	}

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/docs" {
		t.Errorf("/: got %d, Location %q", rec.Code, rec.Header().Get("Location"))
	}

	if code, _ := do(t, h, "GET", "/unknown", ""); code != http.StatusNotFound {
		t.Errorf("/unknown: got %d", code)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	if code, _ := do(t, newTestRouter(t), "GET", "/adduser", ""); code != http.StatusMethodNotAllowed {
		t.Fatalf("got %d", code)
	}
}
