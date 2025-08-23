package firebase

import (
	"context"
	"fmt"
	"strings"

	"errors"

	"cloud.google.com/go/firestore"
	"github.com/theHinneh/budgeting/internal/domain"
	"google.golang.org/api/iterator"
)

type UserRepository struct {
	Firestore *firestore.Client
}

func (f *UserRepository) CreateUser(ctx context.Context, u *domain.User) (*domain.User, error) {
	if u == nil || strings.TrimSpace(u.UID) == "" {
		return nil, fmt.Errorf("invalid user profile")
	}
	_, err := f.Firestore.Collection("users").Doc(u.UID).Set(ctx, map[string]interface{}{
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
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (f *UserRepository) GetUser(ctx context.Context, uid string) (*domain.User, error) {
	dsnap, err := f.Firestore.Collection("users").Doc(uid).Get(ctx)
	if err != nil {
		return nil, err
	}
	var m domain.User
	if err := dsnap.DataTo(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (f *UserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	// Firestore does not directly support querying by email without a specific index.
	// This would require a collection group query or pre-indexing.
	// For simplicity, returning an error for now.
	return nil, fmt.Errorf("GetUserByEmail not directly supported/efficient for Firestore without a specific index. Implement a query on 'Email' field if needed")
}

func (f *UserRepository) UpdateUser(ctx context.Context, u *domain.User) (*domain.User, error) {
	if u == nil || strings.TrimSpace(u.UID) == "" {
		return nil, fmt.Errorf("invalid user profile for update")
	}

	updates := map[string]interface{}{
		"Username":      u.Username,
		"Email":         u.Email,
		"FirstName":     u.FirstName,
		"LastName":      u.LastName,
		"PhoneNumber":   u.PhoneNumber,
		"ProviderID":    u.ProviderID,
		"PhotoURL":      u.PhotoURL,
		"EmailVerified": u.EmailVerified,
		"UpdatedAt":     u.UpdatedAt,
	}

	var firestoreUpdates []firestore.Update
	for k, v := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{Path: k, Value: v})
	}

	_, err := f.Firestore.Collection("users").Doc(u.UID).Update(ctx, firestoreUpdates)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (f *UserRepository) DeleteUser(ctx context.Context, uid string) error {
	_, err := f.Firestore.Collection("users").Doc(uid).Delete(ctx)
	return err
}

func (f *UserRepository) ListAllUserIDs(ctx context.Context) ([]string, error) {
	var userIDs []string
	iter := f.Firestore.Collection("users").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err != nil {
			if errors.Is(err, iterator.Done) {
				break
			}
			return nil, err
		}
		userIDs = append(userIDs, doc.Ref.ID)
	}
	return userIDs, nil
}
