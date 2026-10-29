package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mypocket/backend/internal/usecase"
)

type JWT struct {
	secret []byte
	ttl    time.Duration
}

const (
	claimUserID    = "uid"
	claimEmail     = "em"
	claimName      = "name"
	claimSessionID = "sid"
)

func NewJWT(secret string, ttl time.Duration) *JWT {
	return &JWT{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

func (j *JWT) Generate(userID, email, name, sessionID string) (string, time.Time, error) {
	expiresAt := time.Now().Add(j.ttl)
	claims := jwt.MapClaims{
		claimUserID:    userID,
		claimEmail:     email,
		claimName:      name,
		claimSessionID: sessionID,
		"exp":          expiresAt.Unix(),
		"iat":          time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (j *JWT) Verify(tokenString string) (*usecase.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, errors.New("invalid signing algorithm")
		}
		return j.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, usecase.ErrInvalidToken
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, usecase.ErrInvalidToken
	}
	exp, err := claims.GetExpirationTime()
	if err != nil {
		return nil, usecase.ErrInvalidToken
	}
	uid, _ := claims[claimUserID].(string)
	sid, _ := claims[claimSessionID].(string)
	email, _ := claims[claimEmail].(string)
	name, _ := claims[claimName].(string)
	if uid == "" || sid == "" {
		return nil, usecase.ErrInvalidToken
	}
	return &usecase.TokenClaims{
		UserID:    uid,
		Email:     email,
		Name:      name,
		SessionID: sid,
		ExpiresAt: exp.Time,
	}, nil
}
