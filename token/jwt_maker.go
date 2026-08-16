package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

// JWTMaker is a JSON Web Token maker
type JWTMaker struct {
	secretKey string
}

// NewJWTMakercreate a new JWTMaker
func NewJWTMaker(secretKey string) (Maker, error){
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)

	}
	return &JWTMaker{secretKey}, nil
}

func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username,duration)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(maker.secretKey))
}

	// VerifyToken checks if the token is valid or not
func (maker *JWTMaker)	VerifyToken(token string) (*Payload, error){
	// fmt.Print("jwt_maker","token",token)
	keyFunc := func(token *jwt.Token) (interface{},error){
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey),nil
	}

	payload := &Payload{}

	jwtToken, err := jwt.ParseWithClaims(token, payload, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired){
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !jwtToken.Valid{
		return nil, ErrInvalidToken
	}

	// The following is legacy from the previous Go jwt version
	// payload, ok := jwtToken.Claims.(*Payload)
	// if !ok {
	// 	return nil, ErrInvalidToken
	// }

	return payload, nil
}
