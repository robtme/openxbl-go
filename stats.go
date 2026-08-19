package openxbl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// StatMinutesPlayed is the stat name Xbox uses for a title's playtime. Most, but not
// all, titles report it.
const StatMinutesPlayed = "MinutesPlayed"

// statsBatchSize caps how many stats go into a single batch request. The service's
// real limit is undocumented; raise this if larger requests succeed.
const statsBatchSize = 100

// Stat is a single game statistic. Value is text because a title's stats aren't all
// numeric (ranks and tiers come back as words); use GetInt for the numeric ones.
type Stat struct {
	XUID    string `json:"xuid"`
	SCID    string `json:"scid"`
	TitleID string `json:"titleid"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Value   string `json:"value"`
}

// UnmarshalJSON decodes a stat. The API quotes Value for most stats but not all, so
// bare numbers are kept as their literal text rather than failing the whole response.
func (s *Stat) UnmarshalJSON(data []byte) error {
	type stat Stat // sheds this method, so the embedded decode doesn't recurse

	raw := struct {
		*stat

		Value json.RawMessage `json:"value"`
	}{stat: (*stat)(s)}

	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unmarshaling stat: %w", err)
	}

	if len(raw.Value) == 0 || string(raw.Value) == "null" {
		return nil
	}

	if err := json.Unmarshal(raw.Value, &s.Value); err != nil {
		s.Value = string(raw.Value)
	}

	return nil
}

// GetInt returns the stat's value as an integer, reporting false if it isn't numeric.
func (s *Stat) GetInt() (int64, bool) {
	parsed, err := strconv.ParseInt(s.Value, 10, 64)

	return parsed, err == nil
}

// GetStatsForTitle returns the authenticated account's stats for the given game.
func (c *Client) GetStatsForTitle(ctx context.Context, titleID string) ([]*Stat, error) {
	if titleID == "" {
		return nil, errors.New("missing title ID")
	}

	return c.fetchStats(ctx, http.MethodGet, "achievements/stats/"+titleID, nil)
}

// GetStats returns the named stats for the given user across the given games. Each stat
// is requested per title, so pass one entry per title/stat pair.
func (c *Client) GetStats(ctx context.Context, xboxID string, stats ...StatRequest) ([]*Stat, error) {
	if xboxID == "" {
		return nil, errors.New("missing xbox ID")
	}

	if len(stats) == 0 {
		return nil, errors.New("missing stats")
	}

	var collected []*Stat

	for start := 0; start < len(stats); start += statsBatchSize {
		end := min(start+statsBatchSize, len(stats))

		request := struct {
			XUIDs []string      `json:"xuids"`
			Stats []StatRequest `json:"stats"`
		}{
			XUIDs: []string{xboxID},
			Stats: stats[start:end],
		}

		batch, err := c.fetchStats(ctx, http.MethodPost, "player/stats", request)
		if err != nil {
			return nil, err
		}

		collected = append(collected, batch...)
	}

	return collected, nil
}

// StatRequest names a single stat to fetch for a single game.
type StatRequest struct {
	Name    string `json:"name"`
	TitleID string `json:"titleId"`
}

// GetPlaytime returns the given user's playtime for each of the given games, keyed by
// title ID. Games the service reports no playtime for are omitted.
func (c *Client) GetPlaytime(ctx context.Context, xboxID string, titleIDs ...string) (map[string]time.Duration, error) {
	if len(titleIDs) == 0 {
		return nil, errors.New("missing title IDs")
	}

	requests := make([]StatRequest, 0, len(titleIDs))
	for _, titleID := range titleIDs {
		requests = append(requests, StatRequest{Name: StatMinutesPlayed, TitleID: titleID})
	}

	stats, err := c.GetStats(ctx, xboxID, requests...)
	if err != nil {
		return nil, err
	}

	playtime := make(map[string]time.Duration, len(stats))

	for _, stat := range stats {
		if stat.Name != StatMinutesPlayed {
			continue
		}

		minutes, ok := stat.GetInt()
		if !ok {
			continue
		}

		playtime[stat.TitleID] = time.Duration(minutes) * time.Minute
	}

	return playtime, nil
}

func (c *Client) fetchStats(ctx context.Context, method string, endpoint string, body any) ([]*Stat, error) {
	response := struct {
		StatListsCollection []struct {
			Stats []*Stat `json:"stats"`
		} `json:"statlistscollection"`
	}{}

	if _, err := c.makeRequest(ctx, method, endpoint, body, &response); err != nil {
		return nil, err
	}

	var stats []*Stat
	for _, collection := range response.StatListsCollection {
		stats = append(stats, collection.Stats...)
	}

	return stats, nil
}
