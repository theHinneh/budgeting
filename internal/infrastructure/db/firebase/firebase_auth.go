package firebase

import (
	"context"
	"strings"

	fbAuth "firebase.google.com/go/v4/auth"
)

type FirebaseAuth struct {
	Auth *fbAuth.Client
}

func (f *FirebaseAuth) CreateAuthUser(ctx context.Context, email, password, displayName string, phone *string) (string, error) {
	params := (&fbAuth.UserToCreate{}).
		Email(email).
		Password(password).
		DisplayName(displayName)
	if phone != nil && strings.TrimSpace(*phone) != "" {
		params = params.PhoneNumber(*phone)
	}
	u, err := f.Auth.CreateUser(ctx, params)
	if err != nil {
		return "", err
	}
	return u.UID, nil
}

func (f *FirebaseAuth) GetAuthUser(ctx context.Context, uid string) error {
	_, err := f.Auth.GetUser(ctx, uid)
	return err
}

func (f *FirebaseAuth) UpdateAuthUser(ctx context.Context, uid string, email *string, displayName *string, phone *string) error {
	upd := (&fbAuth.UserToUpdate{})
	changed := false

	if email != nil {
		upd = upd.Email(*email)
		changed = true
	}
	if displayName != nil {
		upd = upd.DisplayName(*displayName)
		changed = true
	}
	if phone != nil && strings.TrimSpace(*phone) != "" {
		upd = upd.PhoneNumber(*phone)
		changed = true
	}

	if !changed {
		return nil // No updates to apply
	}

	_, err := f.Auth.UpdateUser(ctx, uid, upd)
	return err
}

func (f *FirebaseAuth) DeleteAuthUser(ctx context.Context, uid string) error {
	return f.Auth.DeleteUser(ctx, uid)
}

func (f *FirebaseAuth) UpdatePassword(ctx context.Context, uid string, newPassword string) error {
	upd := (&fbAuth.UserToUpdate{}).Password(newPassword)
	_, err := f.Auth.UpdateUser(ctx, uid, upd)
	return err
}

func (f *FirebaseAuth) GeneratePasswordResetLink(ctx context.Context, email string) (string, error) {
	return f.Auth.PasswordResetLink(ctx, email)
}
