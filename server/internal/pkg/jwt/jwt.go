package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"hunjie/internal/config"
)

type Claims struct {
	UserID uint64 `json:"user_id"`
	Role   int8   `json:"role"` // 仅作提示，授权判定一律回查数据库
	jwt.RegisteredClaims
}

// Generate 签发 token
func Generate(userID uint64, role int8) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.AddDate(0, 0, config.C.JWT.ExpireDays)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(config.C.JWT.Secret))
}

// Parse 解析并校验 token
func Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("签名算法不匹配")
			}
			return []byte(config.C.JWT.Secret), nil
		})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("token 无效")
	}
	return claims, nil
}
