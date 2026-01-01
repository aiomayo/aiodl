package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	innertubeAPI = "https://www.youtube.com/youtubei/v1/player"
	apiKey       = "AIzaSyA8eiZmM1FaDVjRy-df2KTyQ_vz_yYM39w"
)

func buildInnertubeRequest(videoID string) innertubeRequest {
	req := innertubeRequest{
		VideoID:        videoID,
		ContentCheckOK: true,
		RacyCheckOK:    true,
	}
	req.Context.Client.ClientName = "ANDROID"
	req.Context.Client.ClientVersion = "19.09.37"
	req.Context.Client.AndroidSDK = 30
	req.Context.Client.HL = "en"
	req.Context.Client.GL = "US"
	req.Context.Client.UserAgent = "com.google.android.youtube/19.09.37 (Linux; U; Android 11) gzip"
	req.Context.Client.OSName = "Android"
	req.Context.Client.OSVersion = "11"
	req.Context.Client.Platform = "MOBILE"
	req.Context.Client.ClientFormFact = "SMALL_FORM_FACTOR"
	req.PlaybackContext.ContentPlaybackContext.HTML5Preference = "HTML5_PREF_WANTS"
	return req
}

func (c *Client) GetVideo(ctx context.Context, urlOrID string) (*Video, error) {
	id, err := ExtractVideoID(urlOrID)
	if err != nil {
		id = urlOrID
	}

	reqBody := buildInnertubeRequest(id)
	url := innertubeAPI + "?key=" + apiKey

	respBody, err := c.httpPost(ctx, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("innertube request: %w", err)
	}

	var pr playerResponse
	if err := json.Unmarshal(respBody, &pr); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if err := checkPlayability(&pr); err != nil {
		return nil, err
	}

	return convertPlayerResponse(&pr), nil
}

func checkPlayability(pr *playerResponse) error {
	switch pr.PlayabilityStatus.Status {
	case "OK":
		return nil
	case "UNPLAYABLE":
		return ErrVideoUnavailable
	case "LOGIN_REQUIRED":
		if pr.VideoDetails.IsPrivate {
			return ErrVideoPrivate
		}
		return ErrAgeRestricted
	case "ERROR":
		if strings.Contains(pr.PlayabilityStatus.Reason, "not found") {
			return ErrVideoNotFound
		}
		return fmt.Errorf("%s", pr.PlayabilityStatus.Reason)
	case "LIVE_STREAM":
		return ErrLiveStream
	default:
		if pr.PlayabilityStatus.LiveStreamAbility.LiveStreamabilityRenderer.VideoID != "" {
			return ErrLiveStream
		}
		return fmt.Errorf("playability status: %s", pr.PlayabilityStatus.Status)
	}
}

func convertPlayerResponse(pr *playerResponse) *Video {
	video := &Video{
		ID:          pr.VideoDetails.VideoID,
		Title:       pr.VideoDetails.Title,
		Description: pr.VideoDetails.ShortDescription,
		Author:      pr.VideoDetails.Author,
		AuthorID:    pr.VideoDetails.ChannelID,
	}

	if d, err := strconv.ParseInt(pr.VideoDetails.LengthSeconds, 10, 64); err == nil {
		video.Duration = time.Duration(d) * time.Second
	}

	if v, err := strconv.ParseInt(pr.VideoDetails.ViewCount, 10, 64); err == nil {
		video.ViewCount = v
	}

	if pd, err := time.Parse("2006-01-02", pr.Microformat.PlayerMicroformatRenderer.PublishDate); err == nil {
		video.PublishDate = pd
	}

	for _, t := range pr.VideoDetails.Thumbnail.Thumbnails {
		video.Thumbnails = append(video.Thumbnails, Thumbnail{
			URL:    t.URL,
			Width:  t.Width,
			Height: t.Height,
		})
	}

	video.formats = make(FormatList, 0, len(pr.StreamingData.Formats)+len(pr.StreamingData.AdaptiveFormats))
	for _, f := range pr.StreamingData.Formats {
		video.formats = append(video.formats, convertFormat(f))
	}
	for _, f := range pr.StreamingData.AdaptiveFormats {
		video.formats = append(video.formats, convertFormat(f))
	}

	return video
}

func convertFormat(f formatRaw) Format {
	format := Format{
		ItagNo:          f.ItagNo,
		MimeType:        f.MimeType,
		Quality:         f.Quality,
		QualityLabel:    f.QualityLabel,
		Bitrate:         f.Bitrate,
		Width:           f.Width,
		Height:          f.Height,
		FPS:             f.FPS,
		AudioQuality:    f.AudioQuality,
		AudioSampleRate: f.AudioSampleRate,
		AudioChannels:   f.AudioChannels,
		URL:             f.URL,
	}

	if f.ContentLength != "" {
		format.ContentLength, _ = strconv.ParseInt(f.ContentLength, 10, 64)
	}

	return format
}
