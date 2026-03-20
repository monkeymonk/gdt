package download

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// DownloadOpts controls optional download behavior.
// Zero-value preserves current behavior (no resume, no mirrors).
type DownloadOpts struct {
	Resume  bool
	Mirrors []string
}

func File(ctx context.Context, url string, dest string, opts DownloadOpts) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	total, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	written := int64(0)
	buf := make([]byte, 32*1024)

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, err := f.Write(buf[:n]); err != nil {
				return err
			}
			written += int64(n)
			if total > 0 {
				pct := float64(written) / float64(total) * 100
				fmt.Fprintf(os.Stderr, "\r  downloading... %.0f%%", pct)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}
	fmt.Fprintln(os.Stderr)
	return nil
}
