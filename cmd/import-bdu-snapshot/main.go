// import-bdu-snapshot downloads the latest БДУ ФСТЭК snapshot from
// bdu-fstec-mirror (gzipped SQLite, ≈50 MB compressed / ≈470 MB decompressed)
// and writes it to disk as ./data/bdu.sqlite for the server to mount as a
// read-only reference catalogue (см. internal/bdu).
//
// Usage:
//
//	import-bdu-snapshot \
//	    --source https://github.com/velvetway/bdu-fstec-mirror/raw/main/data/bdu.sqlite.gz \
//	    --target ./data/bdu.sqlite
package main

import (
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const defaultSource = "https://github.com/velvetway/bdu-fstec-mirror/raw/main/data/bdu.sqlite.gz"

func main() {
	var (
		source = flag.String("source", defaultSource, "URL or local path to bdu.sqlite.gz (or .sqlite)")
		target = flag.String("target", "./data/bdu.sqlite", "Destination path for the unpacked SQLite snapshot")
	)
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	if err := os.MkdirAll(filepath.Dir(*target), 0o755); err != nil {
		log.Fatalf("mkdir target dir: %v", err)
	}

	// Stream-download → optional gunzip → write to target.
	tmp := *target + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		log.Fatalf("create temp: %v", err)
	}
	cleanup := func() { _ = os.Remove(tmp) }

	read, isGzip, closer, err := openSource(ctx, *source)
	if err != nil {
		out.Close()
		cleanup()
		log.Fatalf("open source: %v", err)
	}
	defer closer()

	var src io.Reader = read
	if isGzip {
		gz, err := gzip.NewReader(read)
		if err != nil {
			out.Close()
			cleanup()
			log.Fatalf("gunzip: %v", err)
		}
		defer gz.Close()
		src = gz
	}

	written, err := io.Copy(out, src)
	if err != nil {
		out.Close()
		cleanup()
		log.Fatalf("copy: %v", err)
	}
	if err := out.Close(); err != nil {
		cleanup()
		log.Fatalf("close temp: %v", err)
	}
	if err := os.Rename(tmp, *target); err != nil {
		cleanup()
		log.Fatalf("rename: %v", err)
	}

	fmt.Printf("Wrote %d bytes (%.1f MB) to %s\n", written, float64(written)/1024/1024, *target)
}

// openSource returns a stream for the URL or path, plus whether it should be
// gunzipped before consumption (decided by file extension), and a cleanup func.
func openSource(ctx context.Context, source string) (io.Reader, bool, func(), error) {
	isGzip := strings.HasSuffix(source, ".gz")

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		req, err := http.NewRequestWithContext(ctx, "GET", source, nil)
		if err != nil {
			return nil, false, nil, err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, false, nil, err
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			return nil, false, nil, fmt.Errorf("status %d", resp.StatusCode)
		}
		log.Printf("downloading %s (Content-Length: %s)", source, resp.Header.Get("Content-Length"))
		return resp.Body, isGzip, func() { resp.Body.Close() }, nil
	}

	f, err := os.Open(source)
	if err != nil {
		return nil, false, nil, err
	}
	return f, isGzip, func() { f.Close() }, nil
}
