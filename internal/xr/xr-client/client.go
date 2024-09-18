package xrclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Memonagi/wallet_project/internal/models"
	"github.com/sirupsen/logrus"
)

type Client struct {
	cfg Config
}

type Config struct {
	ServerAddress string
}

const (
	route = "/api/v1/xr?from=%v&to=%v"
)

func New(cfg Config) *Client {
	return &Client{cfg: cfg}
}

var ErrStatus = errors.New("wrong status code")

func (c *Client) GetRate(ctx context.Context, from, to string) (float64, error) {
	address := c.cfg.ServerAddress + fmt.Sprintf(route, from, to)

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
		return 0, ErrStatus
	}

	var response models.XRResponse

	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("xrclient: failed to decode response: %w", err)
	}

	return response.Rate, nil
}
