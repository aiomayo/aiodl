package youtube

import "errors"

var (
	ErrVideoNotFound    = errors.New("video not found")
	ErrVideoPrivate     = errors.New("video is private")
	ErrVideoUnavailable = errors.New("video is unavailable")
	ErrAgeRestricted    = errors.New("video is age-restricted")
	ErrLiveStream       = errors.New("cannot download live streams")
	ErrPlaylistNotFound = errors.New("playlist not found")
	ErrInvalidURL       = errors.New("invalid URL")
	ErrNoFormats        = errors.New("no formats available")
)
