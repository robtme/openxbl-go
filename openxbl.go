package openxbl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const apiURL = "https://api.xbl.io/v2/"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string, timeout time.Duration) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: timeout},
	}
}

// timeLayouts covers the API's inconsistent timestamps. Both are UTC, but a friend's
// addedDateTimeUtc carries a zone while their lastSeenDateTimeUtc omits it. The
// fractional seconds are optional in both.
var timeLayouts = []string{time.RFC3339Nano, "2006-01-02T15:04:05.999999999"}

// Time is a timestamp that decodes whichever of those layouts the API returns. It
// embeds time.Time, so it behaves like one.
type Time struct {
	time.Time
}

func (t *Time) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("unmarshaling timestamp: %w", err)
	}

	if value == "" {
		return nil
	}

	for _, layout := range timeLayouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			t.Time = parsed

			return nil
		}
	}

	return fmt.Errorf("parsing timestamp %q", value)
}
