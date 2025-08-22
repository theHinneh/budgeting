package domain

import (
	"strings"
	"time"
)

type User struct {
	UID           string
	Username      string
	Email         string
	FirstName     string
	LastName      string
	PhoneNumber   *string
	ProviderID    string // e.g., "password", "google", "github"
	PhotoURL      string
	EmailVerified bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewUser(uid, username, email, firstName, lastName string, phoneNumber *string) *User {
	return &User{
		UID:           uid,
		Username:      strings.ToLower(username),
		Email:         strings.ToLower(email),
		FirstName:     firstName,
		LastName:      lastName,
		PhoneNumber:   phoneNumber,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		EmailVerified: true,
	}
}
