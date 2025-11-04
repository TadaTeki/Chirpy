package auth

import (
	"fmt"
	"net/http"
	"strings"
)

func GetAPIKey(headers http.Header) (string, error) {
	val, ok := headers["Authorization"]
	if !ok || len(val) == 0 {
		return "", fmt.Errorf("no authorization header found")
	}

	authHeader := val[0]

	if strings.Index(authHeader, "ApiKey") != 0 {
		return "", fmt.Errorf("no apikey in the header")
	}

	return strings.TrimPrefix(authHeader, "ApiKey "), nil

}
