package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

const browseAPI = "https://www.youtube.com/youtubei/v1/browse"

type browseRequest struct {
	BrowseID string `json:"browseId"`
	Context  struct {
		Client struct {
			ClientName    string `json:"clientName"`
			ClientVersion string `json:"clientVersion"`
			HL            string `json:"hl"`
			GL            string `json:"gl"`
		} `json:"client"`
	} `json:"context"`
}

func buildBrowseRequest(playlistID string) browseRequest {
	req := browseRequest{
		BrowseID: "VL" + playlistID,
	}
	req.Context.Client.ClientName = "WEB"
	req.Context.Client.ClientVersion = "2.20231219.04.00"
	req.Context.Client.HL = "en"
	req.Context.Client.GL = "US"
	return req
}

func (c *Client) GetPlaylist(ctx context.Context, urlOrID string) (*Playlist, error) {
	id, err := ExtractPlaylistID(urlOrID)
	if err != nil {
		id = urlOrID
	}

	reqBody := buildBrowseRequest(id)
	url := browseAPI + "?key=" + apiKey

	respBody, err := c.httpPost(ctx, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("browse request: %w", err)
	}

	var data playlistData
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	playlist := convertPlaylist(id, &data)
	if playlist == nil {
		return nil, ErrPlaylistNotFound
	}

	return playlist, nil
}

func convertPlaylist(id string, data *playlistData) *Playlist {
	playlist := &Playlist{
		ID:          id,
		Title:       data.Metadata.PlaylistMetadataRenderer.Title,
		Description: data.Metadata.PlaylistMetadataRenderer.Description,
	}

	for _, item := range data.Sidebar.PlaylistSidebarRenderer.Items {
		if author := item.PlaylistSidebarSecondaryInfoRenderer.VideoOwner.VideoOwnerRenderer.Title.String(); author != "" {
			playlist.Author = author
			break
		}
	}

	if playlist.Author == "" {
		playlist.Author = data.Header.PlaylistHeaderRenderer.OwnerText.String()
	}

	playlist.Videos = extractPlaylistVideos(data)
	return playlist
}

func extractPlaylistVideos(data *playlistData) []PlaylistEntry {
	var entries []PlaylistEntry

	for _, tab := range data.Contents.TwoColumnBrowseResultsRenderer.Tabs {
		for _, section := range tab.TabRenderer.Content.SectionListRenderer.Contents {
			for _, item := range section.ItemSectionRenderer.Contents {
				for _, video := range item.PlaylistVideoListRenderer.Contents {
					v := video.PlaylistVideoRenderer
					if v.VideoID == "" {
						continue
					}

					entry := PlaylistEntry{
						ID:     v.VideoID,
						Title:  v.Title.String(),
						Author: v.ShortBylineText.String(),
					}

					if idx := v.Index.SimpleText; idx != "" {
						entry.Index, _ = strconv.Atoi(idx)
					}

					if dur := v.LengthSeconds; dur != "" {
						if d, err := strconv.ParseInt(dur, 10, 64); err == nil {
							entry.Duration = time.Duration(d) * time.Second
						}
					}

					entries = append(entries, entry)
				}
			}
		}
	}

	return entries
}
