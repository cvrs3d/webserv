package auth

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
)

func MakeRefreshToken() (string, error) {
    key := make([]byte, 32)
    n, err := rand.Read(key)
    if err != nil {
        return "", fmt.Errorf("failed to read random bytes: %w", err)
    }
    if n != len(key) {
        return "", fmt.Errorf("read %d random bytes; expected %d", n, len(key))
    }
    encoded := hex.EncodeToString(key)
    return encoded, nil
}