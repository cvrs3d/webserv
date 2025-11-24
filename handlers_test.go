package main


import (
	"testing"
)

func TestCleanBody(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "no profane words",
			in:   "Hello world",
			want: "Hello world",
		},
		{
			name: "censors profane words and lowercases",
			in:   "Kerfuffle is bad",
			want: "**** is bad",
		},
		{
			name: "multiple profane words",
			in:   "sharbert fornax ok",
			want: "**** **** ok",
		},
		{
			name: "handles extra spaces",
			in:   "  hello   kerfuffle  ",
			want: "hello ****",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range var
		t.Run(tc.name, func(t *testing.T) {
			got := cleanBody(tc.in)
			if got != tc.want {
				t.Fatalf("cleanBody(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}