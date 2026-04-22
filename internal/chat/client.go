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
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.Id, nil
}

func (c *KimiClient) PreSignUrl(filename string, accessToken, userId string) (string, string, error) {
	bodyMap := map[string]string{
		"action": "file",
		"name":   filename,
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	req, err := http.NewRequest("POST", "https://kimi.moonshot.cn/api/pre-sign-url", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range core.GetFakeHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Msh-Platform", "web")
	req.Header.Set("X-Traffic-Id", userId)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	reader, err := decompressBody(resp)
	if err != nil {
		return "", "", err
	}
	defer reader.Close()

	respBody, _ := io.ReadAll(reader)
	var result struct {
		Url        string `json:"url"`
		ObjectName string `json:"object_name"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", err
	}

	return result.Url, result.ObjectName, nil
}

func (c *KimiClient) UploadToOSS(uploadUrl string, fileData []byte, contentType, accessToken, userId string) error {
	req, err := http.NewRequest("PUT", uploadUrl, bytes.NewBuffer(fileData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", contentType)
	for k, v := range core.GetFakeHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Msh-Platform", "web")
	req.Header.Set("X-Traffic-Id", userId)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *KimiClient) CreateFile(filename, objectName, accessToken, userId string) (string, string, error) {
	bodyMap := map[string]string{
		"type":        "file",
		"name":        filename,
		"object_name": objectName,
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	req, err := http.NewRequest("POST", "https://kimi.moonshot.cn/api/file", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range core.GetFakeHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Msh-Platform", "web")
	req.Header.Set("X-Traffic-Id", userId)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	reader, err := decompressBody(resp)
	if err != nil {
		return "", "", err
	}
	defer reader.Close()

	respBody, _ := io.ReadAll(reader)
	var result struct {
		Id     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", "", err
	}

	return result.Id, result.Status, nil
}

func (c *KimiClient) ParseFile(fileId string, accessToken, userId string) error {
	bodyMap := map[string]interface{}{
		"ids": []string{fileId},
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	req, err := http.NewRequest("POST", "https://kimi.moonshot.cn/api/file/parse_process", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range core.GetFakeHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Msh-Platform", "web")
	req.Header.Set("X-Traffic-Id", userId)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *KimiClient) CreateConversation(model string, name string, accessToken, userId string) (string, error) {
	kimiPlusId := "kimi"
	if len(model) == 20 {
		kimiPlusId = model
	}

	bodyMap := map[string]interface{}{
		"born_from":   "",
		"is_example":  false,
		"kimiplus_id": kimiPlusId,
		"name":        name,
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	req, err := http.NewRequest("POST", "https://kimi.moonshot.cn/api/chat", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range core.GetFakeHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Msh-Platform", "web")
	req.Header.Set("X-Traffic-Id", userId)

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

	respBody, _ := io.ReadAll(reader)
	var result struct {
		Id string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	return result.Id, nil
}

func (c *KimiClient) FakeRequest(accessToken, userId string) {
	endpoints := []string{
		"https://kimi.moonshot.cn/api/user",
		"https://kimi.moonshot.cn/api/chat_1m/user/status",
		"https://kimi.moonshot.cn/api/chat/list",
		"https://kimi.moonshot.cn/api/show_case/list",
	}
	endpoint := endpoints[core.UnixTimestamp()%4] // Simple random

	req, _ := http.NewRequest("GET", endpoint, nil)
	if strings.Contains(endpoint, "list") {
		req, _ = http.NewRequest("POST", endpoint, bytes.NewBuffer([]byte(`{"offset":0,"size":50}`)))
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range core.GetFakeHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Msh-Platform", "web")
	req.Header.Set("X-Traffic-Id", userId)

	resp, err := c.httpClient.Do(req)
	if err == nil {
		resp.Body.Close()
	}
}

func (c *KimiClient) PromptSnippetSubmit(query, accessToken, userId string) {
	body := map[string]interface{}{
		"offset": 0,
		"size":   10,
		"query":  query,
	}
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", "https://kimi.moonshot.cn/api/prompt-snippet/instance", bytes.NewBuffer(bodyBytes))

	req.Header.Set("Content-Type", "application/json")
	for k, v := range core.GetFakeHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Msh-Platform", "web")
	req.Header.Set("X-Traffic-Id", userId)

	resp, err := c.httpClient.Do(req)
	if err == nil {
		resp.Body.Close()
	}
}

func (c *KimiClient) RemoveConversation(convId string, accessToken, userId string) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("https://kimi.moonshot.cn/api/chat/%s", convId), nil)
	if err != nil {
		return err
	}

	for k, v := range core.GetFakeHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-Msh-Platform", "web")
	req.Header.Set("X-Traffic-Id", userId)
	req.Header.Set("Referer", fmt.Sprintf("https://kimi.moonshot.cn/chat/%s", convId))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
