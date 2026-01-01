package youtube

import "time"

type Video struct {
	ID          string
	Title       string
	Description string
	Duration    time.Duration
	Author      string
	AuthorID    string
	ViewCount   int64
	PublishDate time.Time
	Thumbnails  []Thumbnail
	formats     FormatList
}

func (v *Video) Formats() FormatList {
	return v.formats
}

type Thumbnail struct {
	URL    string
	Width  int
	Height int
}

type Format struct {
	ItagNo          int
	MimeType        string
	Quality         string
	QualityLabel    string
	Bitrate         int
	Width           int
	Height          int
	FPS             int
	AudioQuality    string
	AudioSampleRate string
	AudioChannels   int
	ContentLength   int64
	URL             string
}

type Playlist struct {
	ID          string
	Title       string
	Description string
	Author      string
	Videos      []PlaylistEntry
}

type PlaylistEntry struct {
	ID       string
	Title    string
	Author   string
	Duration time.Duration
	Index    int
}

type innertubeRequest struct {
	VideoID         string          `json:"videoId"`
	Context         innertubeClient `json:"context"`
	ContentCheckOK  bool            `json:"contentCheckOk"`
	RacyCheckOK     bool            `json:"racyCheckOk"`
	PlaybackContext struct {
		ContentPlaybackContext struct {
			HTML5Preference string `json:"html5Preference"`
		} `json:"contentPlaybackContext"`
	} `json:"playbackContext"`
}

type innertubeClient struct {
	Client struct {
		ClientName     string `json:"clientName"`
		ClientVersion  string `json:"clientVersion"`
		AndroidSDK     int    `json:"androidSdkVersion"`
		HL             string `json:"hl"`
		GL             string `json:"gl"`
		UserAgent      string `json:"userAgent"`
		OSName         string `json:"osName"`
		OSVersion      string `json:"osVersion"`
		Platform       string `json:"platform"`
		ClientFormFact string `json:"clientFormFactor"`
	} `json:"client"`
}

type playerResponse struct {
	PlayabilityStatus struct {
		Status            string `json:"status"`
		Reason            string `json:"reason"`
		LiveStreamAbility struct {
			LiveStreamabilityRenderer struct {
				VideoID string `json:"videoId"`
			} `json:"liveStreamabilityRenderer"`
		} `json:"liveStreamability"`
	} `json:"playabilityStatus"`
	VideoDetails struct {
		VideoID          string   `json:"videoId"`
		Title            string   `json:"title"`
		LengthSeconds    string   `json:"lengthSeconds"`
		Keywords         []string `json:"keywords"`
		ChannelID        string   `json:"channelId"`
		ShortDescription string   `json:"shortDescription"`
		Thumbnail        struct {
			Thumbnails []struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"thumbnails"`
		} `json:"thumbnail"`
		ViewCount         string `json:"viewCount"`
		Author            string `json:"author"`
		IsLiveContent     bool   `json:"isLiveContent"`
		IsPrivate         bool   `json:"isPrivate"`
		IsUnpluggedCorpus bool   `json:"isUnpluggedCorpus"`
	} `json:"videoDetails"`
	StreamingData struct {
		ExpiresInSeconds string      `json:"expiresInSeconds"`
		Formats          []formatRaw `json:"formats"`
		AdaptiveFormats  []formatRaw `json:"adaptiveFormats"`
	} `json:"streamingData"`
	Microformat struct {
		PlayerMicroformatRenderer struct {
			PublishDate string `json:"publishDate"`
		} `json:"playerMicroformatRenderer"`
	} `json:"microformat"`
}

type formatRaw struct {
	ItagNo          int    `json:"itag"`
	URL             string `json:"url"`
	MimeType        string `json:"mimeType"`
	Bitrate         int    `json:"bitrate"`
	Width           int    `json:"width"`
	Height          int    `json:"height"`
	ContentLength   string `json:"contentLength"`
	Quality         string `json:"quality"`
	QualityLabel    string `json:"qualityLabel"`
	FPS             int    `json:"fps"`
	AudioQuality    string `json:"audioQuality"`
	AudioSampleRate string `json:"audioSampleRate"`
	AudioChannels   int    `json:"audioChannels"`
}

type playlistData struct {
	Contents struct {
		TwoColumnBrowseResultsRenderer struct {
			Tabs []playlistTab `json:"tabs"`
		} `json:"twoColumnBrowseResultsRenderer"`
	} `json:"contents"`
	Metadata struct {
		PlaylistMetadataRenderer struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"playlistMetadataRenderer"`
	} `json:"metadata"`
	Sidebar struct {
		PlaylistSidebarRenderer struct {
			Items []playlistSidebarItem `json:"items"`
		} `json:"playlistSidebarRenderer"`
	} `json:"sidebar"`
	Header struct {
		PlaylistHeaderRenderer struct {
			Title struct {
				SimpleText string `json:"simpleText"`
			} `json:"title"`
			OwnerText textRuns `json:"ownerText"`
		} `json:"playlistHeaderRenderer"`
	} `json:"header"`
}

type playlistTab struct {
	TabRenderer struct {
		Content struct {
			SectionListRenderer struct {
				Contents []playlistSection `json:"contents"`
			} `json:"sectionListRenderer"`
		} `json:"content"`
	} `json:"tabRenderer"`
}

type playlistSection struct {
	ItemSectionRenderer struct {
		Contents []struct {
			PlaylistVideoListRenderer struct {
				Contents []playlistVideo `json:"contents"`
			} `json:"playlistVideoListRenderer"`
		} `json:"contents"`
	} `json:"itemSectionRenderer"`
}

type playlistVideo struct {
	PlaylistVideoRenderer struct {
		VideoID string   `json:"videoId"`
		Title   textRuns `json:"title"`
		Index   struct {
			SimpleText string `json:"simpleText"`
		} `json:"index"`
		ShortBylineText textRuns `json:"shortBylineText"`
		LengthSeconds   string   `json:"lengthSeconds"`
	} `json:"playlistVideoRenderer"`
}

type playlistSidebarItem struct {
	PlaylistSidebarSecondaryInfoRenderer struct {
		VideoOwner struct {
			VideoOwnerRenderer struct {
				Title textRuns `json:"title"`
			} `json:"videoOwnerRenderer"`
		} `json:"videoOwner"`
	} `json:"playlistSidebarSecondaryInfoRenderer"`
}

type textRuns struct {
	Runs []struct {
		Text string `json:"text"`
	} `json:"runs"`
}

func (t textRuns) String() string {
	if len(t.Runs) > 0 {
		return t.Runs[0].Text
	}
	return ""
}
