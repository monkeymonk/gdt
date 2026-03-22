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

	partialPath := dest + ".partial"
	var offset int64

	if opts.Resume {
		if info, err := os.Stat(partialPath); err == nil {
			offset = info.Size()
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	if opts.Resume && offset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", offset))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var f *os.File

	switch {
	case opts.Resume && resp.StatusCode == http.StatusPartialContent:
		// Server supports range; append to partial file.
		f, err = os.OpenFile(partialPath, os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
	case resp.StatusCode == http.StatusOK:
		// Full response — start fresh.
		offset = 0
		if opts.Resume {
			f, err = os.Create(partialPath)
		} else {
			f, err = os.Create(dest)
		}
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	total, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if total > 0 && resp.StatusCode == http.StatusPartialContent {
		total += offset // Content-Length in 206 is remaining bytes
	}
	written := offset
	buf := make([]byte, 32*1024)

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, err := f.Write(buf[:n]); err != nil {
				f.Close()
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
			f.Close()
			return readErr
		}
	}
	fmt.Fprintln(os.Stderr)

	// Close file before rename to avoid "file in use" errors on Windows.
	if err := f.Close(); err != nil {
		return err
	}

	if opts.Resume {
		return os.Rename(partialPath, dest)
	}
	return nil
}
