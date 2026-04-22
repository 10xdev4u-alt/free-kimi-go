package chat

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/princetheprogrammerbtw/kimi-free-api-go/internal/core"
	"github.com/sirupsen/logrus"
)

type ChatCompletionRequest struct {
	Model          string    `json:"model"`
	Messages       []Message `json:"messages"`
	Stream         bool      `json:"stream"`
	UseSearch      bool      `json:"use_search"`
	ConversationId string    `json:"conversation_id"`
	Tools          []Tool    `json:"tools,omitempty"`
}

type Function struct {
	Name       string      `json:"name"`
	Parameters interface{} `json:"parameters,omitempty"`
	Arguments  string      `json:"arguments,omitempty"`
}

type ToolCall struct {
	Index    int      `json:"index"`
	Id       string   `json:"id,omitempty"`
	Type     string   `json:"type,omitempty"`
	Function Function `json:"function"`
}

type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

type Message struct {
	Role       string      `json:"role"`
	Content    interface{} `json:"content"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
	ToolCallId string      `json:"tool_call_id,omitempty"`
}

type OpenAIStreamChunk struct {
	Id      string   `json:"id"`
	Model   string   `json:"model"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
}

type Choice struct {
	Index        int     `json:"index"`
	Delta        Delta   `json:"delta"`
	FinishReason *string `json:"finish_reason"`
}

type Delta struct {
	Role      string     `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func (c *KimiClient) CreateCompletionStream(model string, messages []Message, tools []Tool, accessToken, userId string, useSearch bool, convId string, outputChan chan<- string) error {
	created := time.Now().Unix()
	
	// BEAST MODE: Send initial chunk IMMEDIATELY to prevent CLI timeout
	outputChan <- fmt.Sprintf("data: %s\n\n", mustMarshal(OpenAIStreamChunk{
		Id: convId, Model: model, Object: "chat.completion.chunk", Created: created,
		Choices: []Choice{{Index: 0, Delta: Delta{Role: "assistant"}}},
	}))

	sendMessages := PrepareMessages(messages, tools, convId != "")
	logrus.Infof("[BEAST] Launching Mega-Prompt: %d chars", len(sendMessages[0]["content"].(string)))

	bodyMap := map[string]interface{}{
		"kimiplus_id": "kimi",
		"messages":    sendMessages,
		"refs":        []string{},
		"use_search":  useSearch,
	}
	bodyBytes, _ := json.Marshal(bodyMap)

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://kimi.moonshot.cn/api/chat/%s/completion/stream", convId), bytes.NewBuffer(bodyBytes))
	headers := core.GetFakeHeaders()
	for k, v := range headers {
		if strings.EqualFold(k, "Accept-Encoding") { continue }
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Traffic-Id", userId)
	req.Header.Set("Referer", fmt.Sprintf("https://kimi.moonshot.cn/chat/%s", convId))

	resp, err := c.httpClient.Do(req)
	if err != nil { return err }
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		eb, _ := io.ReadAll(resp.Body)
		logrus.Errorf("[ERROR] Kimi Error %d: %s", resp.StatusCode, string(eb))
		return fmt.Errorf("status %d", resp.StatusCode)
	}

	var bodyReader io.ReadCloser = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, _ := gzip.NewReader(resp.Body)
		if gz != nil { bodyReader = gz; defer gz.Close() }
	}

	scanner := bufio.NewScanner(bodyReader)
	scanner.Buffer(make([]byte, 1024*1024), 20*1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") { continue }
		data := strings.TrimPrefix(line, "data: ")

		var kimiEvent struct {
			Event string `json:"event"`
			Text  string `json:"text"`
			Msg   struct { Type string; Title string; Url string } `json:"msg"`
		}
		if err := json.Unmarshal([]byte(data), &kimiEvent); err != nil { continue }

		if kimiEvent.Event == "cmpl" {
			outputChan <- fmt.Sprintf("data: %s\n\n", mustMarshal(OpenAIStreamChunk{
				Id: convId, Model: model, Object: "chat.completion.chunk", Created: created,
				Choices: []Choice{{Index: 0, Delta: Delta{Content: kimiEvent.Text}}},
			}))
		} else if kimiEvent.Event == "all_done" {
			outputChan <- fmt.Sprintf("data: %s\n\n", mustMarshal(OpenAIStreamChunk{
				Id: convId, Model: model, Object: "chat.completion.chunk", Created: created,
				Choices: []Choice{{Index: 0, Delta: Delta{}, FinishReason: pointerToString("stop")}},
				Usage:   &Usage{PromptTokens: 1, CompletionTokens: 1, TotalTokens: 2},
			}))
			outputChan <- "data: [DONE]\n\n"
			return nil
		}
	}
	return scanner.Err()
}

func mustMarshal(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func pointerToString(s string) *string {
	return &s
}

func (c *KimiClient) UploadFile(fileUrl, accessToken, userId string) (string, error) {
	if strings.Contains(fileUrl, "127.0.0.1") || strings.Contains(fileUrl, "localhost") { return "", fmt.Errorf("SSRF") }
	var fileData []byte
	var contentType string
	var filename = "uploaded_file"
	if strings.HasPrefix(fileUrl, "data:") {
		parts := strings.Split(fileUrl, ",")
		data, _ := base64.StdEncoding.DecodeString(parts[1])
		fileData = data
		contentType = strings.TrimPrefix(strings.Split(parts[0], ";")[0], "data:")
		filename = fmt.Sprintf("%s.%s", core.UUID(false), strings.Split(contentType, "/")[1])
	} else {
		resp, _ := c.httpClient.Get(fileUrl)
		defer resp.Body.Close()
		fileData, _ = io.ReadAll(resp.Body)
		contentType = resp.Header.Get("Content-Type")
	}
	u, o, _ := c.PreSignUrl(filename, accessToken, userId)
	c.UploadToOSS(u, fileData, contentType, accessToken, userId)
	f, s, _ := c.CreateFile(filename, o, accessToken, userId)
	st := core.UnixTimestamp()
	for s != "initialized" {
		if core.UnixTimestamp()-st > 30 { return "", fmt.Errorf("timeout") }
		time.Sleep(2 * time.Second)
		_, s, _ = c.CreateFile(filename, o, accessToken, userId)
	}
	c.ParseFile(f, accessToken, userId)
	return f, nil
}

func PrepareMessages(messages []Message, tools []Tool, isRefConv bool) []map[string]interface{} {
	var history strings.Builder
	for i, msg := range messages {
		msgContent := ""
		switch c := msg.Content.(type) {
		case string: msgContent = c
		case []interface{}:
			for _, part := range c {
				if m, ok := part.(map[string]interface{}); ok { if m["type"] == "text" { msgContent += fmt.Sprintf("%v", m["text"]) } }
			}
		}
		if !isRefConv && i == len(messages)-1 {
			history.WriteString("system: Use tools if needed. Follow Claude Code protocol.\n")
		}
		history.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msgContent))
	}
	return []map[string]interface{}{{"role": "user", "content": history.String()}}
}

func wrapUrlsToTags(content string) string {
	var urlRegex = regexp.MustCompile(`https?://(www\.)?[-a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`)
	return urlRegex.ReplaceAllStringFunc(content, func(url string) string {
		return fmt.Sprintf(`<url id="" type="url" status="" title="" wc="">%s</url>`, url)
	})
}
