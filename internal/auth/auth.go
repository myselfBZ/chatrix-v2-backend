package auth

import "github.com/golang-jwt/jwt/v5"

type Authenticator interface {
	GenerateTokenPair(userID string ,customClaims map[string]any) (*TokenPair, error)
	ValidateAccessToken(tokenString string) (*jwt.Token, error)
	ValidateRefreshToken(tokenString string) (*jwt.Token, error)
	ExtractUserID(token *jwt.Token) (string, error)
	GenerateAccessToken(userID string) (string, error)
}

