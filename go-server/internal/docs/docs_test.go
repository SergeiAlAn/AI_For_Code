package docs

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestFilesInSync(t *testing.T) {
	for name, embedded := range map[string][]byte{"openapi.yaml": spec, "docs.html": page} {
		original, err := os.ReadFile(filepath.Join("..", "..", "..", name))
		if os.IsNotExist(err) {
			t.Skipf("%s not found in the repository root", name)
		}
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(normalize(original), normalize(embedded)) {
			t.Errorf("internal/docs/%s differs from the repository root; copy it: cp ../%s internal/docs/%s", name, name, name)
		}
	}
}

func normalize(b []byte) []byte {
	return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
}
