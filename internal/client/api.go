package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PairingResponse struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
}

type PairingStatus string

const (
	PairingStatusWaiting PairingStatus = "WAITING"
	PairingStatusClaimed PairingStatus = "CLAIMED"
	PairingStatusExpired PairingStatus = "EXPIRED"
)

type PairingStatusResponse struct {
	Status PairingStatus `json:"status"`
	APIKey *string       `json:"apikey"`
}

type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func NewAPIClient(baseURL string, timeout time.Duration) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *APIClient) RequestPairingCode(deviceID string) (*PairingResponse, error) {
	body, _ := json.Marshal(map[string]string{"device_id": deviceID})
	url := fmt.Sprintf("%s/v1/pairing/request", c.BaseURL)
	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("pairing request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pairing request: status %d: %s", resp.StatusCode, string(respBody))
	}
	var pr PairingResponse
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("decode pairing response: %w", err)
	}
	return &pr, nil
}

func (c *APIClient) CheckPairingStatus(deviceID, code string) (*PairingStatusResponse, error) {
	url := fmt.Sprintf("%s/v1/pairing/status?device_id=%s&code=%s", c.BaseURL, deviceID, code)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("check pairing status: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return &PairingStatusResponse{Status: PairingStatusExpired}, nil
	}
	if resp.StatusCode == http.StatusAccepted {
		return &PairingStatusResponse{Status: PairingStatusWaiting}, nil
	}
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("check pairing status: status %d: %s", resp.StatusCode, string(respBody))
	}
	var sr PairingStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, fmt.Errorf("decode pairing status: %w", err)
	}
	return &sr, nil
}

func (c *APIClient) ForwardEvent(targetURL string, payload []byte, eventType string) (int, error) {
	req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(payload))
	if err != nil {
		return 0, fmt.Errorf("create forward request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Event-Type", eventType)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("forward event: %w", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}
