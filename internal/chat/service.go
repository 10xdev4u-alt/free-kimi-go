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
