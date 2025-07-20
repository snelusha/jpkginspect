package exporter

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
)

func WriteJSON(v any, outputPath string) error {
	var (
		w   io.Writer
		err error
	)

	switch outputPath {
	case "", "-":
		w = os.Stdout
	default:
		dir := filepath.Dir(outputPath)
		if err = os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		f, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer f.Close()

		w = f
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return enc.Encode(v)
}
