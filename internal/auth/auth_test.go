package auth

import (
	"log"
	"testing"
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