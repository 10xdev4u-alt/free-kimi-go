package token

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"

	"github.com/princetheprogrammerbtw/kimi-free-api-go/internal/core"
	"github.com/sirupsen/logrus"
)

type TokenManager struct {
	tokens       sync.Map // map[string]*core.TokenInfo
	refreshLocks sync.Map // map[string]*sync.Mutex
	client       MoonshotClient
}

func NewTokenManager(client MoonshotClient) *TokenManager {
	return &TokenManager{
		client: client,
	}
}

func (m *TokenManager) GetToken(authHeader string) (*core.TokenInfo, error) {
	if authHeader == "" {
		return nil, fmt.Errorf("no authorization header")
	}

	// Support comma-separated tokens for rotation
	rawTokens := strings.Split(authHeader, ",")
	selected := rawTokens[rand.Intn(len(rawTokens))]

	// Robustly extract the token, handling "Bearer " prefix case-insensitively
	parts := strings.Fields(selected)
	var refreshToken string
	if len(parts) > 1 && strings.EqualFold(parts[0], "Bearer") {
		refreshToken = parts[1]
	} else {
		refreshToken = parts[0]
	}

	// Check cache
	if val, ok := m.tokens.Load(refreshToken); ok {
		info := val.(*core.TokenInfo)
		if core.UnixTimestamp() < info.RefreshTime {
			return info, nil
		}
	}

	// Need refresh
	return m.refreshToken(refreshToken)
}

func (m *TokenManager) refreshToken(refreshToken string) (*core.TokenInfo, error) {
	// Lock for this specific refresh token
	lockVal, _ := m.refreshLocks.LoadOrStore(refreshToken, &sync.Mutex{})
	lock := lockVal.(*sync.Mutex)
	lock.Lock()
	defer lock.Unlock()

	// Double check after lock
	if val, ok := m.tokens.Load(refreshToken); ok {
		info := val.(*core.TokenInfo)
		if core.UnixTimestamp() < info.RefreshTime {
			return info, nil
		}
	}

	logrus.Infof("Refreshing token: %s...", refreshToken[:10])
	info, err := m.client.RequestToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// 300s expiry (consistent with TS)
	info.RefreshTime = core.UnixTimestamp() + 300
	m.tokens.Store(refreshToken, info)
	logrus.Infof("Successfully refreshed token for user: %s", info.UserId)

	return info, nil
}

// MoonshotClient placeholder - we will define it in internal/chat/client.go
type MoonshotClient interface {
	RequestToken(refreshToken string) (*core.TokenInfo, error)
}
