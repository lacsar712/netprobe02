package twamp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDumpSessionLogFlushRegression guards against the original bug where
// DumpSessionLog wrapped the file in a bufio.Writer but never flushed it,
// leaving the session log empty (body < buffer) or truncated (body > buffer).
// It is independent of runtime_test.go and must not be edited to change the
// persisted bytes.
func TestDumpSessionLogFlushRegression(t *testing.T) {
	t.Run("small body is fully persisted", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "session.log")
		body := "mode=authenticated pad=128\n"
		if err := DumpSessionLog(path, body); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != body {
			t.Fatalf("expected %q, got %q", body, b)
		}
	})

	t.Run("large body spanning multiple buffers is persisted", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "sub", "session.log")
		// Larger than the default bufio.Writer buffer (4096) so a missing
		// flush would leave the trailing chunk unwritten / file truncated.
		body := strings.Repeat("mode=encrypted pad=1472\n", 512)
		if err := DumpSessionLog(path, body); err != nil {
			t.Fatal(err)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(b) != body {
			t.Fatalf("expected %d bytes, got %d bytes (truncated/unflushed)",
				len(body), len(b))
		}
	})
}
