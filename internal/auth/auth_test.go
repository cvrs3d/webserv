package auth


import (
    "testing"
    "time"
	"log"

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