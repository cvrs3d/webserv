package main

import (
	"time"

	"github.com/cvrs3d/webserv/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token	  string	`json:"token,omitempty"`
}

type Chirp struct {
	ID	uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body string `json:"body"`
}

func MapUserDTOToUser(dto database.User) User {
	return User{
		ID:    dto.ID,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
		Email: dto.Email,
	}
}

func MapChirpDTOToChirp(dto database.Chirp) Chirp {
	return Chirp{
		ID: dto.ID,
		UserID: dto.UserID,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
		Body: dto.Body,
	}
}