package auth


import (
	"net/http"
	"strings"
	"fmt"
	)

func GetApiKey(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("no authorization header found")
	}
	fields := strings.Fields(authHeader)
	if len(fields) > 2 || len(fields) <= 1 {
		return "", fmt.Errorf("this is not a valid authorization header")
	}
	if fields[0] != "ApiKey" {
		return "", fmt.Errorf("invalid authorization for api key")
	}
	return fields[1], nil
}
