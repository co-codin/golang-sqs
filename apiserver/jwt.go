package apiserver

import (
	"fmt"
	"go-sqs/config"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var signingMethod = jwt.SigningMethodHS256

type JwtManager struct {
	config *config.Config
}

func NewJwtManager(config *config.Config) *JwtManager {
	return &JwtManager{config: config}
}

type TokenPair struct {
	AccessToken *jwt.Token
	RefreshToken *jwt.Token
}

type CustomClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func (j *JwtManager) Parse(token string) (*jwt.Token, error) {
	parser := jwt.NewParser()

	jwtToken, err := parser.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if t.Method != signingMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte(j.config.JwtSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}
	return jwtToken, nil
}

func (j *JwtManager) IsAccessToken(token string) bool {
	jwtClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}
	if tokenType, ok := jwtClaims["token_type"]; ok {
		return tokenType == "access"
	}

	return false
}

func (j *JwtManager) GenerateTokenPair(userId uuid.UUID) (*TokenPair, error) {
	now := time.Now()
	issuer := "http://" + j.config.ApiServerHost + ":" + j.config.ApiServerPort

	jwtAccessToken := jwt.NewWithClaims(signingMethod, CustomClaims{
		TokenType: "Bearer",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userId.String(),
			Issuer: issuer,
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute * 15)),
			IssuedAt: jwt.NewNumericDate(now),
		},
	})

	key := []byte(j.config.JwtSecret)
	var err error
	jwtAccessToken.Raw, err = jwtAccessToken.SignedString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token")
	}

	jwtRefreshToken := jwt.NewWithClaims(signingMethod, CustomClaims{
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: userId.String(),
			Issuer: issuer,
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute * 15)),
			IssuedAt: jwt.NewNumericDate(now),
		},
	})

	jwtRefreshToken.Raw, err = jwtRefreshToken.SignedString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token")
	}

	return &TokenPair{
		AccessToken: jwtAccessToken,
		RefreshToken: jwtRefreshToken,
	}, nil
}