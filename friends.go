package openxbl

import (
	"context"
	"errors"
	"net/http"
	"time"
)

type Friend struct {
	XUID               string      `json:"xuid"`
	IsFavorite         bool        `json:"isFavorite"`
	IsFollowingCaller  bool        `json:"isFollowingCaller"`
	IsFollowedByCaller bool        `json:"isFollowedByCaller"`
	IsIdentityShared   bool        `json:"isIdentityShared"`
	AddedDateTimeUtc   time.Time   `json:"addedDateTimeUtc"`
	DisplayName        string      `json:"displayName"`
	RealName           string      `json:"realName"`
	DisplayPicURL      string      `json:"displayPicRaw"`
	UseAvatar          bool        `json:"useAvatar"`
	Gamertag           string      `json:"gamertag"`
	GamerScore         string      `json:"gamerScore"`
	XboxOneRep         string      `json:"xboxOneRep"`
	PresenceState      string      `json:"presenceState"`
	PresenceText       string      `json:"presenceText"`
	PresenceDevices    interface{} `json:"presenceDevices"`
	IsBroadcasting     bool        `json:"isBroadcasting"`
	IsCloaked          interface{} `json:"isCloaked"`
	IsQuarantined      bool        `json:"isQuarantined"`
	Suggestion         interface{} `json:"suggestion"`
	Recommendation     interface{} `json:"recommendation"`
	TitleHistory       interface{} `json:"titleHistory"`
	MultiplayerSummary struct {
		InMultiplayerSession int `json:"InMultiplayerSession"`
		InParty              int `json:"InParty"`
	} `json:"multiplayerSummary"`
	RecentPlayer   interface{} `json:"recentPlayer"`
	Follower       interface{} `json:"follower"`
	PreferredColor struct {
		PrimaryColor   string `json:"primaryColor"`
		SecondaryColor string `json:"secondaryColor"`
		TertiaryColor  string `json:"tertiaryColor"`
	} `json:"preferredColor"`
	PresenceDetails        interface{} `json:"presenceDetails"`
	TitlePresence          interface{} `json:"titlePresence"`
	TitleSummaries         interface{} `json:"titleSummaries"`
	PresenceTitleIDs       interface{} `json:"presenceTitleIds"`
	Detail                 interface{} `json:"detail"`
	CommunityManagerTitles interface{} `json:"communityManagerTitles"`
	SocialManager          struct {
		TitleIDs []interface{} `json:"titleIds"`
	} `json:"socialManager"`
	Broadcast         []interface{} `json:"broadcast"`
	TournamentSummary interface{}   `json:"tournamentSummary"`
	Avatar            interface{}   `json:"avatar"`
}

// GetFriends returns all friends.
func (c *Client) GetFriends(ctx context.Context) ([]*Friend, error) {
	response := struct {
		Friends []*Friend `json:"people"`
	}{}

	if _, err := c.makeRequest(ctx, http.MethodGet, "friends", nil, &response); err != nil {
		return nil, err
	}

	if len(response.Friends) == 0 {
		return nil, errors.New("find friends")
	}

	return response.Friends, nil
}

// GetPresence returns the current Presence for the authenticated account.
func (c *Client) GetPresence(ctx context.Context) (*Presence, error) {
	var response Presence

	if _, err := c.makeRequest(ctx, http.MethodGet, "presence", nil, &response); err != nil {
		return nil, err
	}

	if response.ID == "" {
		return nil, errors.New("find presence")
	}

	return &response, nil
}
