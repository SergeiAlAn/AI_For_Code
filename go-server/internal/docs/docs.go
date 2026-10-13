// Package docs serves the OpenAPI specification and the Swagger UI page.
//
// openapi.yaml and docs.html are copies of the files in the repository root:
// go:embed cannot reach outside the module. TestFilesInSync fails if they drift.
package docs

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.yaml
var spec []byte

//go:embed docs.html
var page []byte

func Page(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(page)
}

func Spec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.Write(spec)
}

func RedirectToPage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/docs", http.StatusFound)
}
