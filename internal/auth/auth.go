package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return match, err
	}
	return match, nil
}

func GetBearerToken(headers http.Header) (string, error) {
    headerValue := headers.Get("Authorization")
    if headerValue == "" {
        return "", fmt.Errorf("authorization header missing")
    }

    headerValue = strings.TrimSpace(headerValue)
    parts := strings.Fields(headerValue)
    if len(parts) < 2 {
        return "", fmt.Errorf("authorization header format must be \"Bearer <token>\", authorization token is empty")
    }

    scheme := parts[0]
    if !strings.EqualFold(scheme, "Bearer") {
        return "", fmt.Errorf("authorization scheme must be Bearer, got %q", scheme)
    }

    tokenString := parts[1]
    if tokenString == "" {
        return "", fmt.Errorf("authorization token is empty")
    }

    return tokenString, nil
}