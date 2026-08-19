package openxbl

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Title is a game in a player's history, with whatever decorations the endpoint
// returning it supplies. The achievements endpoints return the same shape.
type Title struct {
	TitleID      string   `json:"titleId"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Devices      []string `json:"devices"`
	DisplayImage string   `json:"displayImage"`
	Achievement  struct {
		CurrentAchievements int     `json:"currentAchievements"`
		TotalAchievements   int     `json:"totalAchievements"`
		CurrentGamerscore   int     `json:"currentGamerscore"`
		TotalGamerscore     int     `json:"totalGamerscore"`
		ProgressPercentage  float64 `json:"progressPercentage"`
	} `json:"achievement"`
	GamePass struct {
		IsGamePass bool `json:"isGamePass"`
	} `json:"gamePass"`
	TitleHistory struct {
		LastTimePlayed time.Time `json:"lastTimePlayed"`
		Visible        bool      `json:"visible"`
		CanHide        bool      `json:"canHide"`
	} `json:"titleHistory"`
	XboxLiveTier string `json:"xboxLiveTier"`
	IsStreamable bool   `json:"isStreamable"`
}

// GetTitles returns the authenticated account's game library. The response carries no
// playtime; use GetPlaytime with the returned title IDs for that.
func (c *Client) GetTitles(ctx context.Context) ([]*Title, error) {
	return c.fetchTitles(ctx, "titles")
}

// GetTitlesForUser returns the given user's game library.
func (c *Client) GetTitlesForUser(ctx context.Context, xboxID string) ([]*Title, error) {
	if xboxID == "" {
		return nil, errors.New("missing xbox ID")
	}

	return c.fetchTitles(ctx, "titles/"+xboxID)
}

func (c *Client) fetchTitles(ctx context.Context, endpoint string) ([]*Title, error) {
	response := struct {
		Titles []*Title `json:"titles"`
	}{}

	if _, err := c.makeRequest(ctx, http.MethodGet, endpoint, nil, &response); err != nil {
		return nil, err
	}

	return response.Titles, nil
}
