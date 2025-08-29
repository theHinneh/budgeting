package firebase

import (
	"context"
	"time"

	fbAuth "firebase.google.com/go/v4/auth"
	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/domain"
)

type FirebaseTokenAuthenticator struct {
	auth *fbAuth.Client
}

func NewFirebaseTokenAuthenticator(auth *fbAuth.Client) ports.TokenAuthenticator {
	return &FirebaseTokenAuthenticator{
		auth: auth,
	}
}

func (f *FirebaseTokenAuthenticator) CreateCustomToken(ctx context.Context, userID string) (string, error) {
	return f.auth.CustomToken(ctx, userID)
}

func (f *FirebaseTokenAuthenticator) VerifyIDToken(ctx context.Context, idToken string) (string, error) {
	token, err := f.auth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", err
	}
	return token.UID, nil
}

func (f *FirebaseTokenAuthenticator) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := f.auth.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	domainUser := &domain.User{
		UID:           user.UID,
		Username:      user.DisplayName,
		Email:         user.Email,
		FirstName:     user.DisplayName,
		LastName:      "",
		PhoneNumber:   &user.PhoneNumber,
		ProviderID:    user.ProviderID,
		PhotoURL:      user.PhotoURL,
		EmailVerified: user.EmailVerified,
		CreatedAt:     time.Unix(user.UserMetadata.CreationTimestamp, 0),
		UpdatedAt:     time.Now(),
	}

	return domainUser, nil
}
