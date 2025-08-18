package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func MakeRefreshToken() (string, error) {
	data := make([]byte, 32)
	if _, err := rand.Read(data); err == nil {
		s := hex.EncodeToString(data)
		return s, nil
	}
	return "", fmt.Errorf("failed to generate refresh token")
}
