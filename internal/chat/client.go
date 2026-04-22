package chat

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/princetheprogrammerbtw/kimi-free-api-go/internal/core"
)

// Helper to handle decompression
func decompressBody(resp *http.Response) (io.ReadCloser, error) {
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		return gzip.NewReader(resp.Body)
	default:
		return resp.Body, nil
	}
}

type KimiClient struct {
	httpClient *http.Client
}

func NewKimiClient() *KimiClient {
	return &KimiClient{
		httpClient: &http.Client{
			Timeout: 600 * time.Second, // 10 Minutes for massive generations
		},
	}
}

func (c *KimiClient) RequestToken(refreshToken string) (*core.TokenInfo, error) {
	req, err := http.NewRequest("GET", "https://kimi.moonshot.cn/api/auth/token/refresh", nil)
	if err != nil {
		return nil, err
	}

	for k, v := range core.GetFakeHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	req.Header.Set("Referer", "https://kimi.moonshot.cn/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader, err := decompressBody(resp)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	body, _ := io.ReadAll(reader)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to refresh token, status: %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// Fetch user info to get userId
	userId, err := c.GetUserInfo(result.AccessToken, refreshToken)
	if err != nil {
		return nil, err
	}

	return &core.TokenInfo{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		UserId:       userId,
	}, nil
}

func (c *KimiClient) GetUserInfo(accessToken, refreshToken string) (string, error) {
	req, err := http.NewRequest("GET", "https://kimi.moonshot.cn/api/user", nil)
	if err != nil {
		return "", err
	}

	for k, v := range core.GetFakeHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Msh-Platform", "web")
	req.Header.Set("X-Traffic-Id", fmt.Sprintf("7%s", core.GenerateRandomString(18, "numeric")))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	reader, err := decompressBody(resp)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	body, _ := io.ReadAll(reader)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("failed to get user info, status: %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Id string `json:"id"`
	}
