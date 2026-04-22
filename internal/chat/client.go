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
