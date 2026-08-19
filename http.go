package openxbl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *Client) makeRequest(ctx context.Context, method string, endpoint string, body any, object any) ([]byte, error) {
	var err error
	var requestBytes []byte

	if body != nil {
		requestBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshalling body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), apiURL+endpoint, bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Authorization", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("unexpected response status: %d: %s", resp.StatusCode, responseBytes)
	}

	// Every endpoint wraps its payload in a standard envelope: {"content": ..., "code": ...},
	// where "content" is the upstream Xbox/Microsoft payload and "code" is its status code.
	if len(responseBytes) > 0 {
		envelope := struct {
			Content json.RawMessage `json:"content"`
			Code    int             `json:"code"`
		}{}

		if err = json.Unmarshal(responseBytes, &envelope); err != nil {
			return nil, fmt.Errorf("unmarshaling response envelope: %w", err)
		}

		// The envelope can report an upstream error even when the gateway returns HTTP 200.
		if envelope.Code != 0 && envelope.Code/100 != 2 {
			return responseBytes, fmt.Errorf("unexpected upstream status: %d: %s", envelope.Code, envelope.Content)
		}

		// Unmarshal the content payload to object if one is provided.
		if object != nil && len(envelope.Content) > 0 {
			if err = json.Unmarshal(envelope.Content, &object); err != nil {
				return nil, fmt.Errorf("unmarshaling response content: %w", err)
			}
		}
	}

	return responseBytes, nil
}
