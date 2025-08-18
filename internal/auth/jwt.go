package auth

import (
	"fmt"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	mySigningKey := []byte(tokenSecret)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    "chirpy",
		Subject:   userID.String(),
		ID:        uuid.NewString(),
	})
	ss, err := token.SignedString(mySigningKey)
	if err != nil {
		return "", err
	}
	return ss, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	var id uuid.UUID
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return id, err
	}
	if token.Valid {
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			subject, err := claims.GetSubject()
			if err != nil {
				return id, fmt.Errorf("%v", err)
			}
			id, err := uuid.Parse(subject)
			if err != nil {
				return id, fmt.Errorf("invalid subject: %v", err)
			}
			return id, nil
		} else {
			return id, fmt.Errorf("unknown claims type, cannot proceed")
		}
	}
	return id, err
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("no authorization header found")
	}
	fields := strings.Fields(authHeader)
	if len(fields) > 2 || len(fields) <= 1 {
		return "", fmt.Errorf("this is not a valid authorization header")
	}
	return fields[1], nil
}
