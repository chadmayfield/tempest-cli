package json

import (
	"bytes"
	"strings"
	"testing"
)

func TestWriteError(t *testing.T) {
	var buf bytes.Buffer
	// Channel is not JSON-serializable
	err := Write(&buf, make(chan int))
	if err == nil {
		t.Error("expected error for un-encodable value")
	}
	if !strings.Contains(err.Error(), "encoding JSON") {
		t.Errorf("expected 'encoding JSON' in error, got: %s", err.Error())
	}
}

func TestWriteSlice(t *testing.T) {
	var buf bytes.Buffer
	data := []map[string]any{
		{"id": 1, "name": "alpha"},
		{"id": 2, "name": "bravo"},
	}
	if err := Write(&buf, data); err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"alpha"`) {
		t.Errorf("missing alpha in output: %s", out)
	}
	if !strings.Contains(out, `"bravo"`) {
		t.Errorf("missing bravo in output: %s", out)
	}
}

func TestWrite(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]any{
		"temp": 22.5,
		"name": "test",
	}
	if err := Write(&buf, data); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, `"temp"`) {
		t.Errorf("output missing temp field: %s", out)
	}
	if !strings.Contains(out, "22.5") {
		t.Errorf("output missing value: %s", out)
	}
	// Should be indented
	if !strings.Contains(out, "  ") {
		t.Errorf("output not indented: %s", out)
	}
}
