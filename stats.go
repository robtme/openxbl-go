package openxbl

import (
	"cmp"
	"context"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"time"
)

// StatMinutesPlayed is the stat name Xbox uses for a title's playtime. Most, but not
// all, titles report it.
const StatMinutesPlayed = "MinutesPlayed"

// StatGroupHero is the stat group a title features on its Xbox game hub, labelled for
// display rather than by internal stat names.
const StatGroupHero = "Hero"

// groupsBatchSize caps how many groups go into a single batch request. The service
// answers eleven or more with a 400 "Invalid Request".
const groupsBatchSize = 10

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

	// GroupProperties says how a title wants a stat shown, and is only set on a stat
	// fetched as part of a group.
	GroupProperties StatDisplay `json:"groupproperties"`
}

// StatDisplay is how a title asks for one of its grouped stats to be presented.
type StatDisplay struct {
	// DisplayFormat is e.g. Integer, Decimal, Percentage, String, ShortTimeSpan or
	// LongTimeSpan (casing varies by title), or empty. A Percentage is a fraction on a
	// Double stat but a whole percent on an Integer one.
	DisplayFormat string `json:"DisplayFormat"`
	DisplayName   string `json:"DisplayName"`
	// DisplaySemantic is Cumulative, Best or Tier.
	DisplaySemantic string `json:"DisplaySemantic"`
	// DisplayUnit qualifies a timespan: Milliseconds, Seconds or Minutes.
	DisplayUnit string `json:"DisplayUnit"`
	// Ordinal is the stat's position within its group, from 0 or 1 depending on the title.
	Ordinal   int    `json:"Ordinal,string"`
	SortOrder string `json:"SortOrder"`
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

	var response statsResponse

	if _, err := c.makeRequest(ctx, http.MethodGet, "achievements/stats/"+titleID, nil, &response); err != nil {
		return nil, err
	}

	return response.stats(), nil
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

	for batch := range slices.Chunk(stats, statsBatchSize) {
		response, err := c.fetchPlayerStats(ctx, xboxID, batch, nil)
		if err != nil {
			return nil, err
		}

		collected = append(collected, response.stats()...)
	}

	return collected, nil
}

// GetHeroStats returns the given user's hero stats per game, keyed by requested title ID
// and sorted by Ordinal. Every requested game is present, with an empty list if it has no
// hero stats. A stat the player hasn't recorded yet has an empty Value.
func (c *Client) GetHeroStats(ctx context.Context, xboxID string, titleIDs ...string) (map[string][]*Stat, error) {
	if xboxID == "" {
		return nil, errors.New("missing xbox ID")
	}

	if len(titleIDs) == 0 {
		return nil, errors.New("missing title IDs")
	}

	// A blank ID would fail its whole batch, taking valid titles down with it.
	if slices.Contains(titleIDs, "") {
		return nil, errors.New("empty title ID")
	}

	hero := make(map[string][]*Stat, len(titleIDs))
	for _, titleID := range titleIDs {
		hero[titleID] = []*Stat{}
	}

	for batch := range slices.Chunk(titleIDs, groupsBatchSize) {
		groups := make([]StatRequest, 0, len(batch))
		for _, titleID := range batch {
			groups = append(groups, StatRequest{Name: StatGroupHero, TitleID: titleID})
		}

		response, err := c.fetchPlayerStats(ctx, xboxID, nil, groups)
		if err != nil {
			return nil, err
		}

		for titleID, stats := range response.groupStats() {
			// ponytail: an echoed ID the caller didn't send is dropped, not remapped; add
			// normalization if the service is ever seen reformatting IDs.
			if _, requested := hero[titleID]; requested {
				hero[titleID] = stats
			}
		}
	}

	return hero, nil
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

// fetchPlayerStats makes a single player/stats request for individual stats, stat
// groups, or both.
func (c *Client) fetchPlayerStats(ctx context.Context, xboxID string, stats []StatRequest, groups []StatRequest) (statsResponse, error) {
	request := struct {
		XUIDs  []string      `json:"xuids"`
		Stats  []StatRequest `json:"stats,omitempty"`
		Groups []StatRequest `json:"groups,omitempty"`
	}{
		XUIDs:  []string{xboxID},
		Stats:  stats,
		Groups: groups,
	}

	var response statsResponse

	_, err := c.makeRequest(ctx, http.MethodPost, "player/stats", request, &response)

	return response, err
}

type statList struct {
	Stats []*Stat `json:"stats"`
}

type statsResponse struct {
	Groups []struct {
		Name                string     `json:"name"`
		TitleID             string     `json:"titleid"`
		StatListsCollection []statList `json:"statlistscollection"`
	} `json:"groups"`
	StatListsCollection []statList `json:"statlistscollection"`
}

// stats flattens the ungrouped stats.
func (r statsResponse) stats() []*Stat {
	var stats []*Stat
	for _, list := range r.StatListsCollection {
		stats = append(stats, list.Stats...)
	}

	return stats
}

// groupStats flattens the grouped stats per title, sorted by Ordinal. A grouped stat
// carries no title ID of its own, so it takes its group's.
func (r statsResponse) groupStats() map[string][]*Stat {
	grouped := make(map[string][]*Stat, len(r.Groups))

	for _, group := range r.Groups {
		// Non-nil so a title without hero stats is an empty list, not null.
		stats := make([]*Stat, 0)

		for _, list := range group.StatListsCollection {
			for _, stat := range list.Stats {
				stat.TitleID = group.TitleID
				stats = append(stats, stat)
			}
		}

		slices.SortStableFunc(stats, func(a, b *Stat) int {
			return cmp.Compare(a.GroupProperties.Ordinal, b.GroupProperties.Ordinal)
		})

		grouped[group.TitleID] = stats
	}

	return grouped
}
