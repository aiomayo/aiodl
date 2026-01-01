package youtube

import (
	"sort"
	"strings"
)

type FormatList []Format

func (fl FormatList) Filter(predicate func(Format) bool) FormatList {
	result := make(FormatList, 0, len(fl))
	for _, f := range fl {
		if predicate(f) {
			result = append(result, f)
		}
	}
	return result
}

func (fl FormatList) FilterVideoOnly() FormatList {
	return fl.Filter(func(f Format) bool { return f.HasVideo() && f.AudioChannels == 0 })
}

func (fl FormatList) FilterAudioOnly() FormatList {
	return fl.Filter(func(f Format) bool { return strings.HasPrefix(f.MimeType, "audio/") })
}

func (fl FormatList) FilterMuxed() FormatList {
	return fl.Filter(func(f Format) bool { return f.IsMuxed() })
}

func (fl FormatList) FilterByQuality(quality string) FormatList {
	return fl.Filter(func(f Format) bool { return f.Quality == quality || f.QualityLabel == quality })
}

func (fl FormatList) FilterByMimeType(mime string) FormatList {
	return fl.Filter(func(f Format) bool { return strings.Contains(f.MimeType, mime) })
}

func (fl FormatList) FilterByItag(itag int) *Format {
	for _, f := range fl {
		if f.ItagNo == itag {
			return &f
		}
	}
	return nil
}

func (fl FormatList) SortByBitrate() FormatList {
	result := make(FormatList, len(fl))
	copy(result, fl)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Bitrate > result[j].Bitrate
	})
	return result
}

func (fl FormatList) SortByResolution() FormatList {
	result := make(FormatList, len(fl))
	copy(result, fl)
	sort.Slice(result, func(i, j int) bool {
		ri := result[i].Width * result[i].Height
		rj := result[j].Width * result[j].Height
		return ri > rj
	})
	return result
}

func (fl FormatList) Best() *Format {
	if len(fl) == 0 {
		return nil
	}
	sorted := fl.SortByBitrate()
	return &sorted[0]
}

func (fl FormatList) BestVideo() *Format {
	video := fl.FilterVideoOnly()
	if len(video) == 0 {
		video = fl.FilterMuxed()
	}
	return video.SortByResolution().Best()
}

func (fl FormatList) BestAudio() *Format {
	return fl.FilterAudioOnly().Best()
}

func (f Format) HasVideo() bool {
	return strings.HasPrefix(f.MimeType, "video/")
}

func (f Format) HasAudio() bool {
	return strings.HasPrefix(f.MimeType, "audio/") || f.AudioChannels > 0
}

func (f Format) IsMuxed() bool {
	return f.HasVideo() && f.AudioChannels > 0
}

func (f Format) Extension() string {
	mime := f.MimeType
	if i := strings.Index(mime, "/"); i >= 0 {
		mime = mime[i+1:]
	}
	if i := strings.Index(mime, ";"); i >= 0 {
		mime = mime[:i]
	}
	return mime
}
