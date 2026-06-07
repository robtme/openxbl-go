package openxbl

import (
	"context"
	"errors"
	"net/http"
	"time"
)

const (
	DVRCaptureTypeClip       = DVRCaptureType("Clip")
	DVRCaptureTypeScreenshot = DVRCaptureType("Screenshot")
	DVRPrivacyBlocked        = DVRPrivacy("Blocked")
	DVRPrivacyEveryone       = DVRPrivacy("Everyone")
	DVRPrivacyPeopleOnMyList = DVRPrivacy("PeopleOnMyList")

	// dvrURITypeDownload identifies the downloadable media URI within a capture's
	// list of URIs (as opposed to thumbnails or other derivatives).
	dvrURITypeDownload = 2
)

type (
	DVRCaptureType string
	DVRPrivacy     string
)

// DVRContentURI is a single addressable URI for a capture's media.
type DVRContentURI struct {
	URI        string    `json:"uri"`
	FileSize   int64     `json:"fileSize"`
	URIType    int       `json:"uriType"`
	Expiration time.Time `json:"expiration"`
}

// DVRThumbnail is a single thumbnail for a capture.
type DVRThumbnail struct {
	URI           string `json:"uri"`
	FileSize      int64  `json:"fileSize"`
	ThumbnailType int    `json:"thumbnailType"`
}

// DVRCapture holds the fields common to clips and screenshots. ID, UploadDate,
// ContentURIs, and Type are normalized by the client from the type-specific keys
// the API returns (e.g. gameClipId/screenshotId, dateRecorded/dateTaken).
type DVRCapture struct {
	ID            string          `json:"-"`         // gameClipId or screenshotId
	TitleID       int             `json:"titleId"`   // Game's ID
	TitleName     string          `json:"titleName"` // Game's name
	XUID          string          `json:"xuid"`
	DeviceType    string          `json:"deviceType"`
	State         int             `json:"state"`
	UserCaption   string          `json:"userCaption"`
	DatePublished time.Time       `json:"datePublished"`
	LastModified  time.Time       `json:"lastModified"`
	UploadDate    time.Time       `json:"-"` // dateRecorded or dateTaken
	Thumbnails    []DVRThumbnail  `json:"thumbnails"`
	ContentURIs   []DVRContentURI `json:"-"` // gameClipUris or screenshotUris
	Type          DVRCaptureType  `json:"-"`
}

// GetDownloadLink returns the downloadable media URI for the capture. It prefers the
// dedicated download URI and falls back to the first available URI, returning an empty
// string if none exist.
func (d *DVRCapture) GetDownloadLink() string {
	for _, contentURI := range d.ContentURIs {
		if contentURI.URIType == dvrURITypeDownload {
			return contentURI.URI
		}
	}

	if len(d.ContentURIs) > 0 {
		return d.ContentURIs[0].URI
	}

	return ""
}

type Clip struct {
	DVRCapture

	GameClipID        string          `json:"gameClipId"`
	DateRecorded      time.Time       `json:"dateRecorded"`
	DurationInSeconds int             `json:"durationInSeconds"`
	GameClipURIs      []DVRContentURI `json:"gameClipUris"`
}

func (c *Client) DeleteDVRClip(ctx context.Context, id string) error {
	if _, err := c.makeRequest(ctx, http.MethodGet, "dvr/gameclips/delete/"+id, nil, nil); err != nil {
		return err
	}

	return nil
}

func (c *Client) GetDVRClips(ctx context.Context, continuationToken string) ([]*Clip, string, error) {
	response := struct {
		Clips      []*Clip `json:"gameClips"`
		PagingInfo struct {
			ContinuationToken string `json:"continuationToken"`
		} `json:"pagingInfo"`
	}{}

	// If a continuation token (their version of pagination) is supplied, pass it to the API.
	endpoint := "dvr/gameclips"
	if continuationToken != "" {
		endpoint += "?continuationToken=" + continuationToken
	}

	if _, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, "", err
	}

	if len(response.Clips) == 0 {
		return nil, "", errors.New("find clips")
	}

	for _, clip := range response.Clips {
		clip.ID = clip.GameClipID
		clip.UploadDate = clip.DateRecorded
		clip.ContentURIs = clip.GameClipURIs
		clip.Type = DVRCaptureTypeClip
	}

	return response.Clips, response.PagingInfo.ContinuationToken, nil
}

type Screenshot struct {
	DVRCapture

	ScreenshotID     string          `json:"screenshotId"`
	DateTaken        time.Time       `json:"dateTaken"`
	ResolutionHeight int             `json:"resolutionHeight"`
	ResolutionWidth  int             `json:"resolutionWidth"`
	ScreenshotURIs   []DVRContentURI `json:"screenshotUris"`
}

func (c *Client) GetDVRScreenshots(ctx context.Context, continuationToken string) ([]*Screenshot, string, error) {
	response := struct {
		Screenshots []*Screenshot `json:"screenshots"`
		PagingInfo  struct {
			ContinuationToken string `json:"continuationToken"`
		} `json:"pagingInfo"`
	}{}

	// If a continuation token (their version of pagination) is supplied, pass it to the API.
	endpoint := "dvr/screenshots"
	if continuationToken != "" {
		endpoint += "?continuationToken=" + continuationToken
	}

	if _, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, "", err
	}

	if len(response.Screenshots) == 0 {
		return nil, "", errors.New("find screenshots")
	}

	for _, screenshot := range response.Screenshots {
		screenshot.ID = screenshot.ScreenshotID
		screenshot.UploadDate = screenshot.DateTaken
		screenshot.ContentURIs = screenshot.ScreenshotURIs
		screenshot.Type = DVRCaptureTypeScreenshot
	}

	return response.Screenshots, response.PagingInfo.ContinuationToken, nil
}

func (c *Client) SetDVRPrivacy(ctx context.Context, privacy DVRPrivacy) error {
	switch privacy {
	case DVRPrivacyBlocked, DVRPrivacyEveryone, DVRPrivacyPeopleOnMyList:
	default:
		return errors.New("invalid privacy type")
	}

	request := struct {
		Privacy string `json:"value"`
	}{
		Privacy: string(privacy),
	}

	if _, err := c.makeRequest(ctx, http.MethodPost, "dvr/privacy", request, nil); err != nil {
		return err
	}

	return nil
}
