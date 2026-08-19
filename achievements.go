package openxbl

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	AchievementProgressAchieved   = "Achieved"
	AchievementProgressInProgress = "InProgress"
	AchievementProgressNotStarted = "NotStarted"

	// achievementRewardGamerscore identifies the gamerscore entry within an
	// achievement's reward list (as opposed to in-game or art rewards).
	achievementRewardGamerscore = "Gamerscore"
)

// AchievementRequirement tracks progress towards a single unlock condition.
type AchievementRequirement struct {
	ID            string `json:"id"`
	Current       string `json:"current"`
	Target        string `json:"target"`
	OperationType string `json:"operationType"`
	ValueType     string `json:"valueType"`

	RuleParticipationType string `json:"ruleParticipationType"`
}

// AchievementMediaAsset is an image associated with an achievement (typically its icon).
type AchievementMediaAsset struct {
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

// AchievementReward is a single reward granted by an achievement. Gamerscore is the
// common case; titles may also grant in-game items or art.
type AchievementReward struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	ValueType   string `json:"valueType"`
}

// Achievement is a single Xbox One (or later) achievement.
type Achievement struct {
	ID                string `json:"id"`
	ServiceConfigID   string `json:"serviceConfigId"`
	Name              string `json:"name"`
	TitleAssociations []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"titleAssociations"`
	ProgressState string `json:"progressState"`
	Progression   struct {
		Requirements []AchievementRequirement `json:"requirements"`
		TimeUnlocked time.Time                `json:"timeUnlocked"`
	} `json:"progression"`
	MediaAssets       []AchievementMediaAsset `json:"mediaAssets"`
	Platforms         []string                `json:"platforms"`
	IsSecret          bool                    `json:"isSecret"`
	Description       string                  `json:"description"`
	LockedDescription string                  `json:"lockedDescription"`
	ProductID         string                  `json:"productId"`
	AchievementType   string                  `json:"achievementType"`
	ParticipationType string                  `json:"participationType"`
	Rewards           []AchievementReward     `json:"rewards"`
	EstimatedTime     string                  `json:"estimatedTime"`
	Deeplink          string                  `json:"deeplink"`
	IsRevoked         bool                    `json:"isRevoked"`

	// Rarity is how many players hold the achievement, e.g. "Rare" at 0.14%.
	Rarity struct {
		CurrentCategory   string  `json:"currentCategory"`
		CurrentPercentage float64 `json:"currentPercentage"`
	} `json:"rarity"`
}

// GetGamerscore returns the gamerscore awarded by the achievement, or 0 if it awards none.
func (a *Achievement) GetGamerscore() int {
	for _, reward := range a.Rewards {
		if reward.Type != achievementRewardGamerscore {
			continue
		}

		gamerscore, err := strconv.Atoi(reward.Value)
		if err != nil {
			return 0
		}

		return gamerscore
	}

	return 0
}

// IsUnlocked reports whether the achievement has been earned.
func (a *Achievement) IsUnlocked() bool {
	return a.ProgressState == AchievementProgressAchieved
}

// X360Achievement is a single Xbox 360 achievement. The legacy contract is flatter
// than Achievement and uses numeric identifiers.
type X360Achievement struct {
	ID                int       `json:"id"`
	TitleID           int64     `json:"titleId"`
	Name              string    `json:"name"`
	Sequence          int       `json:"sequence"`
	Flags             int       `json:"flags"`
	UnlockedOnline    bool      `json:"unlockedOnline"`
	Unlocked          bool      `json:"unlocked"`
	IsSecret          bool      `json:"isSecret"`
	Description       string    `json:"description"`
	LockedDescription string    `json:"lockedDescription"`
	Gamerscore        int       `json:"gamerscore"`
	ImageID           int64     `json:"imageId"`
	Type              int       `json:"type"`
	Platform          int       `json:"platform"`
	TimeUnlocked      time.Time `json:"timeUnlocked"`
}

// GetAchievementTitles returns the authenticated account's achievement progress, grouped by title.
func (c *Client) GetAchievementTitles(ctx context.Context) ([]*Title, error) {
	return c.fetchTitles(ctx, "achievements")
}

// GetAchievementTitlesForUser returns the given user's achievement progress, grouped by title.
func (c *Client) GetAchievementTitlesForUser(ctx context.Context, xboxID string) ([]*Title, error) {
	if xboxID == "" {
		return nil, errors.New("missing xbox ID")
	}

	return c.fetchTitles(ctx, "achievements/player/"+xboxID)
}

// GetAchievementsForTitle returns the authenticated account's achievements for the given game,
// along with a continuation token for the next page (empty when there are no more pages).
func (c *Client) GetAchievementsForTitle(ctx context.Context, titleID string, continuationToken string) ([]*Achievement, string, error) {
	if titleID == "" {
		return nil, "", errors.New("missing title ID")
	}

	// If a continuation token (their version of pagination) is supplied, pass it to the API.
	endpoint := "achievements/title/" + titleID
	if continuationToken != "" {
		endpoint += "?continuationToken=" + url.QueryEscape(continuationToken)
	}

	return c.fetchAchievements(ctx, endpoint)
}

// GetAchievementsForUser returns the given user's achievements for the given game.
func (c *Client) GetAchievementsForUser(ctx context.Context, xboxID string, titleID string) ([]*Achievement, error) {
	if xboxID == "" {
		return nil, errors.New("missing xbox ID")
	}

	if titleID == "" {
		return nil, errors.New("missing title ID")
	}

	achievements, _, err := c.fetchAchievements(ctx, "achievements/player/"+xboxID+"/"+titleID)

	return achievements, err
}

// GetAchievementsForTitles returns the authenticated account's achievements for the given games.
func (c *Client) GetAchievementsForTitles(ctx context.Context, titleIDs ...string) ([]*Achievement, error) {
	if len(titleIDs) == 0 {
		return nil, errors.New("missing title IDs")
	}

	achievements, _, err := c.fetchAchievements(ctx, "achievements/"+strings.Join(titleIDs, ","))

	return achievements, err
}

// GetX360AchievementsForTitle returns every achievement defined by the given Xbox 360 game.
func (c *Client) GetX360AchievementsForTitle(ctx context.Context, xboxID string, titleID string) ([]*X360Achievement, error) {
	if xboxID == "" {
		return nil, errors.New("missing xbox ID")
	}

	if titleID == "" {
		return nil, errors.New("missing title ID")
	}

	return c.fetchX360Achievements(ctx, "achievements/player/"+xboxID+"/title/"+titleID)
}

// GetX360AchievementsForUser returns the given user's achievements for the given Xbox 360 game.
func (c *Client) GetX360AchievementsForUser(ctx context.Context, xboxID string, titleID string) ([]*X360Achievement, error) {
	if xboxID == "" {
		return nil, errors.New("missing xbox ID")
	}

	if titleID == "" {
		return nil, errors.New("missing title ID")
	}

	return c.fetchX360Achievements(ctx, "achievements/x360/"+xboxID+"/title/"+titleID)
}

// fetchAchievements fetches an achievement list, returning the continuation token for
// the endpoints that paginate. An empty list is valid, so it isn't treated as an error.
func (c *Client) fetchAchievements(ctx context.Context, endpoint string) ([]*Achievement, string, error) {
	response := struct {
		Achievements []*Achievement `json:"achievements"`
		PagingInfo   struct {
			ContinuationToken string `json:"continuationToken"`
		} `json:"pagingInfo"`
	}{}

	if _, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, "", err
	}

	return response.Achievements, response.PagingInfo.ContinuationToken, nil
}

func (c *Client) fetchX360Achievements(ctx context.Context, endpoint string) ([]*X360Achievement, error) {
	response := struct {
		Achievements []*X360Achievement `json:"achievements"`
	}{}

	if _, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, err
	}

	return response.Achievements, nil
}
