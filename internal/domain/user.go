package domain

import (
	"strings"
	"time"
)

type User struct {
	UID           string    `json:"uid"`
	Username      string    `json:"username"`
	Email         string    `json:"email"`
	FirstName     string    `json:"firstname"`
	LastName      string    `json:"lastname"`
	PhoneNumber   *string   `json:"phone_number"`
	ProviderID    string    `json:"provider_id,omitempty"` // e.g., "password", "google", "github"
	PhotoURL      string    `json:"photo_url,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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
