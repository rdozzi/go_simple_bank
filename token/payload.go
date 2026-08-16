package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Different types of error retruned by the VerifyToken function
var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

// Payload contains the payload data of the token
type Payload struct {
	ID uuid.UUID `json:"id"`
	Username string `json:"username"`

	jwt.RegisteredClaims
	// IssuedAt time.Time `json:"issued_at"`
	// ExpiredAt time.Time `json:"expired_at"`
}

// NewPayload creates a new token payload with a specific username and duration
func NewPayload(username string, duration time.Duration) (*Payload, error){
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	payload := &Payload{
		ID: tokenID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
		
	}

	return payload, nil
}

// No longer needed in V5 of jwt golang
// Valid checks if the token payload is valid or not
// func (payload *Payload) Valid() error {
// 	if time.Now().After(payload.RegisteredClaims.ExpiresAt){
// 		return ErrExpiredToken
// 	}

// 	return nil
// }