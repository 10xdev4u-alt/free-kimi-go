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

