package main

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type UserClaims struct {
	jwt.RegisteredClaims
	Ip string `json:"ip"`
	Id string `json:"id"`
}

const (
	AccessTokenInspireTime  = 24 * time.Hour
	RefreshTokenInspireTime = 7 * 24 * time.Hour
)

func createTokens(ip string, id string, secret string) (string, string, error) {
	accessClaims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenInspireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Ip: ip,
		Id: id,
	}

	accessToken, err := generateToken(secret, accessClaims)
	if err != nil {
		return "", "", err
	}

	refreshClaims := UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenInspireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Ip: ip,
		Id: id,
	}

	refreshToken, err := generateToken(secret, refreshClaims)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func generateToken(secret string, claims jwt.Claims) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func validateToken(secret string, tokenString string, claims *UserClaims) (*jwt.Token, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// Проверяем, что используется ожидаемый метод подписи
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("неожиданный метод подписи: %v", t.Header["alg"])
		}
		// Возвращаем секретный ключ для jwt токена, в формате []byte, совпадающий с ключом, использованным для подписи ранее
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("недействительный токен")
	}

	return parsedToken, nil
}
