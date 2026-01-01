package youtube

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const chunkSize = 10 * 1024 * 1024

type ProgressFunc func(downloaded, total int64)

type DownloadOptions struct {
	Itag      int
	MimeType  string
	Quality   string
	AudioOnly bool
}

func (c *Client) Download(ctx context.Context, video *Video, opts DownloadOptions, progress ProgressFunc) (io.ReadCloser, error) {
	format := selectFormat(video.formats, opts)
	if format == nil {
		return nil, ErrNoFormats
	}
	return c.DownloadFormat(ctx, format, progress)
}

func (c *Client) DownloadFormat(ctx context.Context, format *Format, progress ProgressFunc) (io.ReadCloser, error) {
	if format.URL == "" {
		return nil, ErrNoFormats
	}

	contentLength := format.ContentLength
	if contentLength == 0 {
		length, err := c.getContentLength(ctx, format.URL)
		if err == nil && length > 0 {
			contentLength = length
		}
	}

	if contentLength > chunkSize {
		return c.downloadChunked(ctx, format.URL, contentLength, progress)
	}

	return c.downloadSimple(ctx, format.URL, contentLength, progress)
}

func (c *Client) downloadSimple(ctx context.Context, url string, total int64, progress ProgressFunc) (io.ReadCloser, error) {
	resp, err := c.httpGet(ctx, url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}

	if total == 0 {
		total = resp.ContentLength
	}

	if progress == nil {
		return resp.Body, nil
	}

	return &progressReader{
		reader:   resp.Body,
		total:    total,
		progress: progress,
	}, nil
}

func (c *Client) downloadChunked(ctx context.Context, url string, total int64, progress ProgressFunc) (io.ReadCloser, error) {
	pr, pw := io.Pipe()

	go func() {
		defer func() { _ = pw.Close() }()
		var downloaded int64

		for start := int64(0); start < total; start += chunkSize {
			if ctx.Err() != nil {
				pw.CloseWithError(ctx.Err())
				return
			}

			end := start + chunkSize - 1
			if end >= total {
				end = total - 1
			}

			chunk, err := c.downloadRange(ctx, url, start, end)
			if err != nil {
				pw.CloseWithError(err)
				return
			}

			n, err := io.Copy(pw, chunk)
			_ = chunk.Close()
			if err != nil {
				pw.CloseWithError(err)
				return
			}

			downloaded += n
			if progress != nil {
				progress(downloaded, total)
			}
		}
	}()

	return pr, nil
}

func (c *Client) downloadRange(ctx context.Context, url string, start, end int64) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("http status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (c *Client) getContentLength(ctx context.Context, url string) (int64, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Set("User-Agent", userAgent)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("http status %d", resp.StatusCode)
	}

	if cl := resp.Header.Get("Content-Length"); cl != "" {
		return strconv.ParseInt(cl, 10, 64)
	}

	return resp.ContentLength, nil
}

func selectFormat(formats FormatList, opts DownloadOptions) *Format {
	if len(formats) == 0 {
		return nil
	}

	if opts.Itag > 0 {
		return formats.FilterByItag(opts.Itag)
	}

	if opts.AudioOnly {
		return formats.FilterAudioOnly().Best()
	}

	if opts.Quality != "" {
		if f := formats.FilterByQuality(opts.Quality).Best(); f != nil {
			return f
		}
	}

	if opts.MimeType != "" {
		if f := formats.FilterByMimeType(opts.MimeType).Best(); f != nil {
			return f
		}
	}

	if f := formats.FilterMuxed().Best(); f != nil {
		return f
	}

	return formats.Best()
}

type progressReader struct {
	reader     io.ReadCloser
	downloaded int64
	total      int64
	progress   ProgressFunc
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.downloaded += int64(n)
		if pr.progress != nil {
			pr.progress(pr.downloaded, pr.total)
		}
	}
	return n, err
}

func (pr *progressReader) Close() error {
	return pr.reader.Close()
}
