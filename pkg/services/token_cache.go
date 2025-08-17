package services

import (
	"time"
)

type CachedToken struct {
	UserID    int64
	Role      byte
	ExpiresAt time.Time
}

type TokenCache struct {
	tokens map[string]*CachedToken
}

var tokenCache *TokenCache

func GetTokenCache() *TokenCache {
	if tokenCache == nil {
		tokenCache = &TokenCache{
			tokens: make(map[string]*CachedToken),
		}
	}
	return tokenCache
}

func (tc *TokenCache) AddToken(tokenString string, userID int64, role byte, expiresAt time.Time) {
	tc.tokens[tokenString] = &CachedToken{
		UserID:    userID,
		Role:      role,
		ExpiresAt: expiresAt,
	}
}

func (tc *TokenCache) GetToken(tokenString string) (*CachedToken, bool) {
	token, exists := tc.tokens[tokenString]
	if !exists {
		return nil, false
	}

	if time.Now().After(token.ExpiresAt) {
		delete(tc.tokens, tokenString)
		return nil, false
	}

	return token, true
}
