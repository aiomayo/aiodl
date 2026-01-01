package adapter

import (
	"context"
	"io"
)

type MediaType string

const (
	MediaTypeVideo    MediaType = "video"
	MediaTypeAudio    MediaType = "audio"
	MediaTypePlaylist MediaType = "playlist"
)

type MediaInfo struct {
	ID       string
	Title    string
	Duration int
	Type     MediaType
	URL      string
	Platform string
	Formats  []Format
	Items    []MediaInfo
}

type Format struct {
	ID        string
	Extension string
	Quality   string
	FileSize  int64
	Bitrate   int
	Width     int
	Height    int
}

func (f *Format) IsVideo() bool {
	return f.Width > 0 || f.Height > 0
}

type DownloadOptions struct {
	FormatID  string
	Quality   string
	AudioOnly bool
}

type DownloadProgress struct {
	Downloaded int64
	Total      int64
}

type ProgressFunc func(DownloadProgress)

type Adapter interface {
	Name() string
	Matches(url string) bool
	GetInfo(ctx context.Context, url string) (*MediaInfo, error)
	Download(ctx context.Context, info *MediaInfo, opts DownloadOptions, progress ProgressFunc) (io.ReadCloser, error)
}
