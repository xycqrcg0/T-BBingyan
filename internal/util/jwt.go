package util

import (
	"BBingyan/internal/config"
	"BBingyan/internal/log"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type JWTClaim struct {
	Auth       string
	Permission int //0用户;1管理员
	jwt.RegisteredClaims
}

func GenerateJWT(em string, per int) (string, error) {
	claims := &JWTClaim{
		Auth:             em,
		Permission:       per,
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.Config.JWT.Exp) * time.Minute))},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(config.Config.JWT.Key))

	return "Bearer " + signedToken, err
}

func ParseJWT(signedToken string) (*JWTClaim, error) {
	if len(signedToken) > 7 && signedToken[:7] == "Bearer " {
		signedToken = signedToken[7:]
	} else {
		return nil, errors.New("invalid")
	}

	jwtClaims := &JWTClaim{}
	token, err := jwt.ParseWithClaims(signedToken, jwtClaims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return "", errors.New("invalid token")
		}
		return []byte(config.Config.JWT.Key), nil
	})
	if err != nil {
		log.Debugf("err:%v", err)
		return nil, err
	}

	if token.Valid {
		return jwtClaims, nil
	}

	return nil, errors.New("expired token")
}
