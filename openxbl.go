package openxbl

import (
	"net/http"
	"time"
)

const url = "https://api.xbl.io/v2/"

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
