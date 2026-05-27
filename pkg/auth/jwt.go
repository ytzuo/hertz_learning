package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type JWTManager struct {
	secret []byte
	ttl    time.Duration
	issuer string
}

type Claims struct {
	Subject string `json:"sub"`
	Name    string `json:"name,omitempty"`
	Email   string `json:"email,omitempty"`
	Issuer  string `json:"iss"`
	Issued  int64  `json:"iat"`
	Expires int64  `json:"exp"`
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

func NewJWTManager(secret string, ttl time.Duration, issuer string) *JWTManager {
	return &JWTManager{
		secret: []byte(secret),
		ttl:    ttl,
		issuer: issuer,
	}
}

func (m *JWTManager) Sign(subject string, name string, email string) (string, Claims, error) {
	now := time.Now()
	claims := Claims{
		Subject: subject,
		Name:    name,
		Email:   email,
		Issuer:  m.issuer,
		Issued:  now.Unix(),
		Expires: now.Add(m.ttl).Unix(),
	}

	headerPart, err := encodeJSON(jwtHeader{Algorithm: "HS256", Type: "JWT"})
	if err != nil {
		return "", Claims{}, err
	}

	claimsPart, err := encodeJSON(claims)
	if err != nil {
		return "", Claims{}, err
	}

	unsigned := headerPart + "." + claimsPart
	signature := sign(unsigned, m.secret)

	return unsigned + "." + signature, claims, nil
}

func (m *JWTManager) Verify(token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}

	unsigned := parts[0] + "." + parts[1]
	expectedSignature := sign(unsigned, m.secret)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return Claims{}, ErrInvalidToken
	}

	var header jwtHeader
	if err := decodeJSON(parts[0], &header); err != nil {
		return Claims{}, ErrInvalidToken
	}
	if header.Algorithm != "HS256" || header.Type != "JWT" {
		return Claims{}, ErrInvalidToken
	}

	var claims Claims
	if err := decodeJSON(parts[1], &claims); err != nil {
		return Claims{}, ErrInvalidToken
	}
	if claims.Issuer != m.issuer || claims.Subject == "" {
		return Claims{}, ErrInvalidToken
	}
	if time.Now().Unix() >= claims.Expires {
		return Claims{}, ErrExpiredToken
	}

	return claims, nil
}

func BearerToken(authorization string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(authorization, prefix) {
		return "", fmt.Errorf("%w: missing bearer prefix", ErrInvalidToken)
	}

	token := strings.TrimSpace(strings.TrimPrefix(authorization, prefix))
	if token == "" {
		return "", ErrInvalidToken
	}

	return token, nil
}

func encodeJSON(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodeJSON(part string, v any) error {
	data, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func sign(unsigned string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
