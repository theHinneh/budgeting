package firebase

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"

	"github.com/theHinneh/budgeting/internal/application/ports"
)

type FirebaseTokenGenerator struct{}

func NewFirebaseTokenGenerator() ports.TokenGenerator {
	return &FirebaseTokenGenerator{}
}

func (f *FirebaseTokenGenerator) GenerateSecureToken() (string, error) {

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}

func (f *FirebaseTokenGenerator) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
