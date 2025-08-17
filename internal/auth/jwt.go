package auth

import (
	"fmt"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"time"
	"errors"
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
	})
	switch {
	case token.Valid:
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
	case errors.Is(err, jwt.ErrTokenMalformed):
		return id, fmt.Errorf("token is malformed")
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		// Invalid signature
		return id, fmt.Errorf("invalid signature")
	case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
		// Token is either expired or not active yet
		return id, fmt.Errorf("token has expired or still inactive")
	default:
		return id, fmt.Errorf("unknown error: %v", err)
	}

}
