package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
    // convert secret to []byte explicitly
    secretKey := []byte(tokenSecret)

    claims := jwt.RegisteredClaims{
        Issuer:    "chirpy",
        IssuedAt:  jwt.NewNumericDate(time.Now()),
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
        Subject:   userID.String(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := token.SignedString(secretKey)
    if err != nil {
        return "", err
    }
    return signed, nil
}

type MyCustomClaims struct {
	jwt.RegisteredClaims
}	

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &MyCustomClaims{}
	t, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return []byte(tokenSecret), nil
		})
	
	if err != nil {
		return uuid.UUID{}, err
	}

	if !t.Valid {
		return uuid.UUID{}, fmt.Errorf("token is invalid")
	}

	user := claims.Subject

	return uuid.Parse(user)
}