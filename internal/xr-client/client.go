package xrclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Client struct {
	Rate float64
}

const port = 2607

func New() *Client {
	return &Client{}
}

var errStatus = errors.New("wrong status code")

func (c *Client) GetRate(ctx context.Context, from, to string) (float64, error) {
	address := fmt.Sprintf("http://localhost:%d/xr?from=%v&to=%v", port, from, to)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("xrclient: failed to send request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Warnf("xrclient: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return 0, errStatus
	}

	var responseBody struct {
		Rate float64 `json:"rate"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return 0, fmt.Errorf("xrclient: failed to decode response: %w", err)
	}

	c.Rate = responseBody.Rate

	return c.Rate, nil
}
