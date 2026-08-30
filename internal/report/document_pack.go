package report

import (
	"archive/zip"
	"bytes"
	"fmt"
	"time"
)

// DocumentPackEntry — одна составляющая пакета документации.
type DocumentPackEntry struct {
	Filename string
	PDF      []byte
}

// PackDocuments склеивает несколько PDF в один ZIP.
// Имя файла в архиве — entry.Filename.
func PackDocuments(entries []DocumentPackEntry) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range entries {
		if len(e.PDF) == 0 {
			continue
		}
		header := &zip.FileHeader{
			Name:     e.Filename,
			Method:   zip.Deflate,
			Modified: time.Now(),
		}
		w, err := zw.CreateHeader(header)
		if err != nil {
			return nil, fmt.Errorf("zip create %s: %w", e.Filename, err)
		}
		if _, err := w.Write(e.PDF); err != nil {
			return nil, fmt.Errorf("zip write %s: %w", e.Filename, err)
		}
	}
	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("zip close: %w", err)
	}
	return buf.Bytes(), nil
}
