package auth

import (
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)


func TestHashingPasswords(t *testing.T) {
	tests := []struct{
		name string
		password string
	}{
		{
			name: "1.",
			password: "passwrod",
		},
		{
			name: "2.",
			password: "21321n4rd",
		},
		{
			name: "3.",
			password: "12mdmd",
		},
		{
			name: "4.",
			password: "!@$@!$@((!JDN@!@BSNN!@NJB DB ))",
		},
	}
	
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			hashed, _ := HashPassword(tc.password)
			if flag, _ := CheckPasswordHash(tc.password, hashed); flag != true {
				log.Fatalf("Hashing failed: first hash(%s) != password hash (%s)", hashed, tc.password)
			}
		})
	}
}

func TestMakeAndValidateJWT_Success(t *testing.T) {
    secret := "super-secret-key-123"
    userID := uuid.New()
    expiresIn := time.Minute * 10

    token, err := MakeJWT(userID, secret, expiresIn)
    require.NoError(t, err, "MakeJWT should not error")

    gotUserID, err := ValidateJWT(token, secret)
    require.NoError(t, err, "ValidateJWT should succeed")
    require.Equal(t, userID, gotUserID, "ValidateJWT should return the same userID that was signed")
}

func TestValidateJWT_WrongSecret(t *testing.T) {
    secret := "super-secret-key-123"
    wrongSecret := "wrong-key"
    userID := uuid.New()
    expiresIn := time.Minute * 10

    token, err := MakeJWT(userID, secret, expiresIn)
    require.NoError(t, err)

    _, err = ValidateJWT(token, wrongSecret)
    require.Error(t, err, "ValidateJWT should error if secret is wrong")
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
    secret := "super-secret-key-123"
    userID := uuid.New()
    // expire immediately
    expiresIn := time.Millisecond * 1

    token, err := MakeJWT(userID, secret, expiresIn)
    require.NoError(t, err)

    // wait for it to expire
    time.Sleep(time.Millisecond * 5)

    _, err = ValidateJWT(token, secret)
    require.Error(t, err, "ValidateJWT should error for expired token")
}

func TestGetBearerToken(t *testing.T) {
    tests := []struct {
        name        string
        headerValue string
        wantToken   string
        wantErr     bool
        errContains string
    }{
        {
            name:        "missing Authorization header",
            headerValue: "",
            wantToken:   "",
            wantErr:     true,
            errContains: "authorization header missing",
        },
        {
            name:        "empty header value",
            headerValue: "   ",
            wantToken:   "",
            wantErr:     true,
            errContains: "authorization header format",
        },
        {
            name:        "wrong scheme",
            headerValue: "Basic someToken",
            wantToken:   "",
            wantErr:     true,
            errContains: "authorization scheme must be Bearer",
        },
        {
            name:        "only scheme no token",
            headerValue: "Bearer",
            wantToken:   "",
            wantErr:     true,
            errContains: "authorization header format",
        },
        {
            name:        "token empty after scheme",
            headerValue: "Bearer  ",
            wantToken:   "",
            wantErr:     true,
            errContains: "authorization token is empty",
        },
        {
            name:        "valid bearer token lowercase scheme",
            headerValue: "bearer abc.def.ghi",
            wantToken:   "abc.def.ghi",
            wantErr:     false,
        },
        {
            name:        "valid bearer token uppercase scheme",
            headerValue: "BEARER xyz123",
            wantToken:   "xyz123",
            wantErr:     false,
        },
        {
            name:        "valid bearer token with extra spaces",
            headerValue: "  Bearer   jwt.token.string  ",
            wantToken:   "jwt.token.string",
            wantErr:     false,
        },
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            headers := http.Header{}
            if tc.headerValue != "" {
                headers.Set("Authorization", tc.headerValue)
            }

            tok, err := GetBearerToken(headers)
            if tc.wantErr {
                require.Error(t, err)
                require.Contains(t, err.Error(), tc.errContains)
            } else {
                require.NoError(t, err)
                require.Equal(t, tc.wantToken, tok)
            }
        })
    }
}