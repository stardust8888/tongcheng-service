package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"tongcheng/conf"
	"tongcheng/utils/constant"
)

// Claims JWT声明结构体
type Claims struct {
	UserID   int    `json:"userID"`
	WxOpenID string `json:"wxOpenID"`
	IsAdmin  int    `json:"isAdmin"`
	jwt.RegisteredClaims
}

// JWT结构体
type JWT struct {
	Secret     []byte
	ExpireTime time.Duration
}

// 定义错误
var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("expired token")
	ErrEmptySecret  = errors.New("secret is empty")
)

// New 创建JWT实例
func New(secret string, expireTime time.Duration) (*JWT, error) {
	if secret == "" {
		return nil, ErrEmptySecret
	}

	return &JWT{
		Secret:     []byte(secret),
		ExpireTime: expireTime,
	}, nil
}

// GenerateToken 生成JWT Token
func (j *JWT) GenerateToken(user constant.SimpleUser) (string, error) {
	// 设置过期时间
	expireAt := time.Now().Add(j.ExpireTime)

	// 创建声明
	claims := Claims{
		UserID:   user.ID,
		WxOpenID: user.WxOpenID,
		IsAdmin:  user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    conf.GetConfiguration().App.Name,
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	tokenString, err := token.SignedString(j.Secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析并验证JWT Token
func (j *JWT) ParseToken(tokenString string) (*Claims, error) {
	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return j.Secret, nil
	})

	if err != nil {
		// 检查是否是过期错误
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	// 验证token
	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// 获取声明
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// RefreshToken 刷新JWT Token
func (j *JWT) RefreshToken(tokenString string) (string, error) {
	// 解析现有token
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 生成新token
	simple := constant.SimpleUser{
		ID:       claims.UserID,
		WxOpenID: claims.WxOpenID,
		IsAdmin:  claims.IsAdmin,
	}
	newToken, err := j.GenerateToken(simple)
	if err != nil {
		return "", err
	}

	return newToken, nil
}

// VerifyToken 验证JWT Token
func (j *JWT) VerifyToken(tokenString string) bool {
	_, err := j.ParseToken(tokenString)
	return err == nil
}

// GetUserIDFromToken 从token中获取用户ID
func (j *JWT) GetUserIDFromToken(tokenString string) (int, error) {
	claims, err := j.ParseToken(tokenString)
	if err != nil {
		return 0, err
	}

	return claims.UserID, nil
}

// IsTokenExpired 检查token是否过期
func (j *JWT) IsTokenExpired(tokenString string) bool {
	_, err := j.ParseToken(tokenString)
	return errors.Is(err, ErrExpiredToken)
}
