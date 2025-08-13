package models

import "time"

type User struct {
	UID         string  `json:"uid"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	FirstName   string  `json:"firstname"`
	LastName    string  `json:"lastname"`
	PhoneNumber *string `json:"phone_number,omitempty"`
	Password    string  `json:"password,omitempty" firestore:"-" gorm:"-"`

	// Optional metadata
	ProviderID    string    `json:"provider_id,omitempty"` // e.g., "password", "google.com"
	PhotoURL      string    `json:"photo_url,omitempty"`
	EmailVerified bool      `json:"email_verified,omitempty"`
	CreatedAt     time.Time `json:"created_at,omitempty"`
	UpdatedAt     time.Time `json:"updated_at,omitempty"`
}

func NewUser(uid, username, email, firstName, lastName string, phone *string) *User {
	return &User{
		UID:         uid,
		Username:    username,
		Email:       email,
		FirstName:   firstName,
		LastName:    lastName,
		PhoneNumber: phone,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
}
