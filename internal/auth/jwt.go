package auth

import (
	"fmt"
	"maps"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthenticator struct {
	accessSecret  string
	refreshSecret string
	aud           string
	iss           string
}

func NewJWTAuthenticator(accessSecret, refreshSecret, aud, iss string) Authenticator {
	return &JWTAuthenticator{
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		aud:           aud,
		iss:           iss,
	}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string	`json:"refresh_token"`
}

func (a *JWTAuthenticator) GenerateTokenPair(userID string, customClaims map[string]any) (*TokenPair, error) {
	accessClaims := jwt.MapClaims{
		"sub": userID,
		"aud": a.aud,
		"iss": a.iss,
		"exp": time.Now().Add(15 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"type": "access",
	}

	maps.Copy(accessClaims, customClaims)
	
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(a.accessSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}
	
	// Refresh token (long-lived: 7 days)
	refreshClaims := jwt.MapClaims{
		"sub": userID,
		"aud": a.aud,
		"iss": a.iss,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"type": "refresh",
	}
	
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(a.refreshSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}
	
	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

// ValidateAccessToken validates the access token
func (a *JWTAuthenticator) ValidateAccessToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		
		// Verify it's an access token
		if claims, ok := t.Claims.(jwt.MapClaims); ok {
			if tokenType, exists := claims["type"]; !exists || tokenType != "access" {
				return nil, fmt.Errorf("invalid token type")
			}
		}
		
		return []byte(a.accessSecret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(a.aud),
		jwt.WithIssuer(a.iss),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}

func (a *JWTAuthenticator) ValidateRefreshToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method %v", t.Header["alg"])
		}
		
		// Verify it's a refresh token
		if claims, ok := t.Claims.(jwt.MapClaims); ok {
			if tokenType, exists := claims["type"]; !exists || tokenType != "refresh" {
				return nil, fmt.Errorf("invalid token type")
			}
		}
		
		return []byte(a.refreshSecret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(a.aud),
		jwt.WithIssuer(a.iss),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}


func (a *JWTAuthenticator) GenerateAccessToken(userID string) (string, error) {
	accessClaims := jwt.MapClaims{
		"sub": userID,
		"aud": a.aud,
		"iss": a.iss,
		"exp": time.Now().Add(15 * time.Hour).Unix(),
		"iat": time.Now().Unix(),
		"type": "access",
	}
	
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(a.accessSecret))

	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return accessTokenString, nil
}


func (a *JWTAuthenticator) ExtractUserID(token *jwt.Token) (string, error) {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}
	
	userID, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("invalid user ID in token")
	}
	
	return userID, nil
}

