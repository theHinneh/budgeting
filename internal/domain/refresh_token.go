package domain

import (
	"time"
)

type RefreshToken struct {
	ID         string     `json:"id" firestore:"id"`
	UserID     string     `json:"user_id" firestore:"user_id"`
	TokenHash  string     `json:"token_hash" firestore:"token_hash"`
	IsRevoked  bool       `json:"is_revoked" firestore:"is_revoked"`
	ExpiresAt  time.Time  `json:"expires_at" firestore:"expires_at"`
	CreatedAt  time.Time  `json:"created_at" firestore:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty" firestore:"revoked_at,omitempty"`
	DeviceInfo string     `json:"device_info,omitempty" firestore:"device_info,omitempty"`
	IPAddress  string     `json:"ip_address,omitempty" firestore:"ip_address,omitempty"`
	UserAgent  string     `json:"user_agent,omitempty" firestore:"user_agent,omitempty"`
}
