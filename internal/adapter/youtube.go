package adapter

import (
	"context"
	"io"
	"strconv"

	"github.com/aiomayo/aiodl/youtube"
)

type YouTubeAdapter struct {
	client *youtube.Client
}

func NewYouTubeAdapter(opts ...youtube.Option) *YouTubeAdapter {
	return &YouTubeAdapter{client: youtube.NewClient(opts...)}
}

func (a *YouTubeAdapter) Name() string { return "youtube" }

func (a *YouTubeAdapter) Matches(url string) bool {
	return youtube.IsYouTubeURL(url)
}

func (a *YouTubeAdapter) GetInfo(ctx context.Context, url string) (*MediaInfo, error) {
	if _, err := youtube.ExtractPlaylistID(url); err == nil {
		return a.getPlaylistInfo(ctx, url)
	}
	return a.getVideoInfo(ctx, url)
}

func (a *YouTubeAdapter) getVideoInfo(ctx context.Context, url string) (*MediaInfo, error) {
	video, err := a.client.GetVideo(ctx, url)
	if err != nil {
		return nil, err
	}

	info := &MediaInfo{
		ID:       video.ID,
		Title:    video.Title,
		Duration: int(video.Duration.Seconds()),
		Type:     MediaTypeVideo,
		URL:      url,
		Platform: "youtube",
	}

	for _, f := range video.Formats() {
		info.Formats = append(info.Formats, Format{
			ID:        strconv.Itoa(f.ItagNo),
			Extension: f.Extension(),
			Quality:   f.QualityLabel,
			FileSize:  f.ContentLength,
			Bitrate:   f.Bitrate,
			Width:     f.Width,
			Height:    f.Height,
		})
	}

	return info, nil
}

func (a *YouTubeAdapter) getPlaylistInfo(ctx context.Context, url string) (*MediaInfo, error) {
	playlist, err := a.client.GetPlaylist(ctx, url)
	if err != nil {
		return nil, err
	}

	info := &MediaInfo{
		ID:       playlist.ID,
		Title:    playlist.Title,
		Type:     MediaTypePlaylist,
		URL:      url,
		Platform: "youtube",
	}

	for _, entry := range playlist.Videos {
		info.Items = append(info.Items, MediaInfo{
			ID:       entry.ID,
			Title:    entry.Title,
			Duration: int(entry.Duration.Seconds()),
			Type:     MediaTypeVideo,
			URL:      "https://www.youtube.com/watch?v=" + entry.ID,
			Platform: "youtube",
		})
	}

	return info, nil
}

func (a *YouTubeAdapter) Download(ctx context.Context, info *MediaInfo, opts DownloadOptions, progress ProgressFunc) (io.ReadCloser, error) {
	video, err := a.client.GetVideo(ctx, info.URL)
	if err != nil {
		return nil, err
	}

	ytOpts := youtube.DownloadOptions{
		Itag:      parseItag(opts.FormatID),
		Quality:   opts.Quality,
		AudioOnly: opts.AudioOnly,
	}

	var ytProgress youtube.ProgressFunc
	if progress != nil {
		ytProgress = func(downloaded, total int64) {
			progress(DownloadProgress{Downloaded: downloaded, Total: total})
		}
	}

	return a.client.Download(ctx, video, ytOpts, ytProgress)
}

func parseItag(formatID string) int {
	if formatID == "" {
		return 0
	}
	itag, _ := strconv.Atoi(formatID)
	return itag
}

var _ Adapter = (*YouTubeAdapter)(nil)
