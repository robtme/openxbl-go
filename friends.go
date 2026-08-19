package openxbl

import (
	"context"
	"errors"
	"net/http"
)

// FriendPresenceDetail describes what a friend is doing on a single device. The API
// capitalises these keys, unlike the rest of the payload.
type FriendPresenceDetail struct {
	Device           string `json:"Device"`
	DeviceSubType    string `json:"DeviceSubType"`
	GameplayType     string `json:"GameplayType"`
	IsBroadcasting   bool   `json:"IsBroadcasting"`
	IsGame           bool   `json:"IsGame"`
	IsPrimary        bool   `json:"IsPrimary"`
	PresenceText     string `json:"PresenceText"`
	RichPresenceText string `json:"RichPresenceText"`
	State            string `json:"State"`
	TitleID          string `json:"TitleId"`
	TitleType        string `json:"TitleType"`
}

// FriendLinkedAccount is an external account shown on a friend's profile, such as
// Steam, Discord, or Twitch.
type FriendLinkedAccount struct {
	NetworkName      string `json:"networkName"`
	DisplayName      string `json:"displayName"`
	Deeplink         string `json:"deeplink"`
	IsFamilyFriendly bool   `json:"isFamilyFriendly"`
	ShowOnProfile    bool   `json:"showOnProfile"`
}

type Friend struct {
	XUID                 string                 `json:"xuid"`
	IsFavorite           bool                   `json:"isFavorite"`
	IsFollowingCaller    bool                   `json:"isFollowingCaller"`
	IsFollowedByCaller   bool                   `json:"isFollowedByCaller"`
	IsIdentityShared     bool                   `json:"isIdentityShared"`
	IsQuarantined        bool                   `json:"isQuarantined"`
	AddedDateTimeUtc     Time                   `json:"addedDateTimeUtc"`
	LastSeenDateTimeUtc  Time                   `json:"lastSeenDateTimeUtc"`
	DisplayName          string                 `json:"displayName"`
	RealName             string                 `json:"realName"`
	DisplayPicURL        string                 `json:"displayPicRaw"`
	IsXbox360Gamerpic    bool                   `json:"isXbox360Gamerpic"`
	ShowUserAsAvatar     string                 `json:"showUserAsAvatar"`
	Gamertag             string                 `json:"gamertag"`
	ModernGamertag       string                 `json:"modernGamertag"`
	ModernGamertagSuffix string                 `json:"modernGamertagSuffix"`
	UniqueModernGamertag string                 `json:"uniqueModernGamertag"`
	GamerScore           string                 `json:"gamerScore"`
	XboxOneRep           string                 `json:"xboxOneRep"`
	ColorTheme           string                 `json:"colorTheme"`
	PreferredFlag        string                 `json:"preferredFlag"`
	PreferredPlatforms   []string               `json:"preferredPlatforms"`
	LinkedAccounts       []FriendLinkedAccount  `json:"linkedAccounts"`
	IsBroadcasting       bool                   `json:"isBroadcasting"`
	PresenceState        string                 `json:"presenceState"`
	PresenceText         string                 `json:"presenceText"`
	PresenceDetails      []FriendPresenceDetail `json:"presenceDetails"`
	MultiplayerSummary   struct {
		InParty int `json:"inParty"`
	} `json:"multiplayerSummary"`
	PreferredColor struct {
		PrimaryColor   string `json:"primaryColor"`
		SecondaryColor string `json:"secondaryColor"`
		TertiaryColor  string `json:"tertiaryColor"`
		ColorURI       string `json:"colorUri"`
	} `json:"preferredColor"`
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
