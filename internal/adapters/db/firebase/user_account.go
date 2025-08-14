package firebase

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/firestore"
	fbAuth "firebase.google.com/go/v4/auth"
	"github.com/theHinneh/budgeting/internal/core/models"
	"github.com/theHinneh/budgeting/internal/core/ports"
)

var _ ports.UserAccountPort = (*Database)(nil)

// CreateAuthUser creates a Firebase Auth user and returns UID.
func (d *Database) CreateAuthUser(ctx context.Context, email, password, displayName string, phone *string) (string, error) {
	params := (&fbAuth.UserToCreate{}).
		Email(email).
		Password(password).
		DisplayName(displayName)
	if phone != nil && strings.TrimSpace(*phone) != "" {
		params = params.PhoneNumber(*phone)
	}
	u, err := d.Auth.CreateUser(ctx, params)
	if err != nil {
		return "", err
	}
	return u.UID, nil
}

// GetAuthUser returns nil if the user exists, otherwise an error.
func (d *Database) GetAuthUser(ctx context.Context, uid string) error {
	_, err := d.Auth.GetUser(ctx, uid)
	return err
}

// UpdateAuthUser updates email/phone/display name for an auth user.
func (d *Database) UpdateAuthUser(ctx context.Context, uid string, email *string, displayName *string, phone *string) error {
	var upd *fbAuth.UserToUpdate
	if email != nil {
		if upd == nil {
			upd = &fbAuth.UserToUpdate{}
		}
		upd = upd.Email(*email)
	}
	if displayName != nil {
		if upd == nil {
			upd = &fbAuth.UserToUpdate{}
		}
		upd = upd.DisplayName(*displayName)
	}
	if phone != nil {
		if upd == nil {
			upd = &fbAuth.UserToUpdate{}
		}
		if strings.TrimSpace(*phone) != "" {
			upd = upd.PhoneNumber(*phone)
		}
	}
	if upd == nil {
		return nil
	}
	_, err := d.Auth.UpdateUser(ctx, uid, upd)
	return err
}

func (d *Database) DeleteAuthUser(ctx context.Context, uid string) error {
	return d.Auth.DeleteUser(ctx, uid)
}

func (d *Database) SaveProfile(ctx context.Context, u *models.User) error {
	if u == nil || strings.TrimSpace(u.UID) == "" {
		return fmt.Errorf("invalid user profile")
	}
	_, err := d.Firestore.Collection("users").Doc(u.UID).Set(ctx, map[string]interface{}{
		"UID":           u.UID,
		"Username":      u.Username,
		"Email":         u.Email,
		"FirstName":     u.FirstName,
		"LastName":      u.LastName,
		"PhoneNumber":   u.PhoneNumber,
		"ProviderID":    u.ProviderID,
		"PhotoURL":      u.PhotoURL,
		"EmailVerified": u.EmailVerified,
		"CreatedAt":     u.CreatedAt,
		"UpdatedAt":     u.UpdatedAt,
	})
	return err
}

func (d *Database) GetProfile(ctx context.Context, uid string) (*models.User, error) {
	dsnap, err := d.Firestore.Collection("users").Doc(uid).Get(ctx)
	if err != nil {
		return nil, err
	}
	var m models.User
	if err := dsnap.DataTo(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (d *Database) UpdateProfile(ctx context.Context, uid string, updates map[string]interface{}) error {
	var ups []firestore.Update
	for k, v := range updates {
		ups = append(ups, firestore.Update{Path: k, Value: v})
	}
	_, err := d.Firestore.Collection("users").Doc(uid).Update(ctx, ups)
	return err
}

func (d *Database) DeleteProfile(ctx context.Context, uid string) error {
	_, err := d.Firestore.Collection("users").Doc(uid).Delete(ctx)
	return err
}
