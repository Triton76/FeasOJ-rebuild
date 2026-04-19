// Layer: Security (安全基础设施层)
// Responsibility: JWT Token生成与解析、Claims定义
// Dependency: 不依赖业务层，是通用工具层，被 Usecase 和 Middleware 使用
package security

import (
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"uid"`
	Role   string `json:"role"`
	Status string `json:"status"`
	jwt.RegisteredClaims
}

func GenerateToken(secret, issuer string, ttl time.Duration, userID, role, status string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: userID,
		Role:   role,
		Status: status,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(secret, tokenString string) (Claims, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return Claims{}, errors.New("empty token")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return Claims{}, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	if claims.UserID == "" {
		return Claims{}, errors.New("invalid claims")
	}

	return *claims, nil
}
