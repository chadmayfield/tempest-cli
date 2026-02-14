package json

import (
	"encoding/json"
	"fmt"
	"io"
)

// Write encodes v as indented JSON to w.
func Write(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}
