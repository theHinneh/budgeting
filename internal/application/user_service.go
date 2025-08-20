package application

import (
	"context"
	"strings"
	"time"

	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/domain"
)

type UserService struct {
	accounts ports.UserAccountPort
}

func NewUserService(accounts ports.UserAccountPort) *UserService {
	return &UserService{accounts: accounts}
}

var _ ports.UserServicePort = (*UserService)(nil)

func (s *UserService) CreateUser(ctx context.Context, in ports.CreateUserInput) (string, error) {
	// Basic validation
	if strings.TrimSpace(in.Username) == "" || strings.TrimSpace(in.Email) == "" || strings.TrimSpace(in.FirstName) == "" || strings.TrimSpace(in.LastName) == "" {
		return "", ErrValidation
	}

	var uid string
	password := strings.TrimSpace(in.Password)
	if password != "" {
		// Create auth user
		display := strings.TrimSpace(strings.TrimSpace(in.FirstName + " " + in.LastName))
		u, err := s.accounts.CreateAuthUser(ctx, in.Email, password, display, in.PhoneNumber)
		if err != nil {
			return "", err
		}
		uid = u
	} else {
		uid = strings.TrimSpace(in.UID)
		if uid == "" {
			return "", ErrValidation
		}
		// Verify auth user exists
		if err := s.accounts.GetAuthUser(ctx, uid); err != nil {
			return "", err
		}
	}

	// Build and store profile
	user := domain.NewUser(uid, in.Username, in.Email, in.FirstName, in.LastName, in.PhoneNumber)
	if err := s.accounts.SaveProfile(ctx, user); err != nil {
		return "", err
	}
	return uid, nil
}

func (s *UserService) GetUser(ctx context.Context, uid string) (*domain.User, error) {
	return s.accounts.GetProfile(ctx, strings.TrimSpace(uid))
}

func (s *UserService) UpdateUser(ctx context.Context, uid string, in ports.UpdateUserInput) (*domain.User, error) {
	uid = strings.TrimSpace(uid)
	updates := map[string]interface{}{
		"UpdatedAt": time.Now().UTC(),
	}
	if in.Username != nil {
		updates["Username"] = *in.Username
	}
	if in.Email != nil {
		updates["Email"] = *in.Email
	}
	if in.FirstName != nil {
		updates["FirstName"] = *in.FirstName
	}
	if in.LastName != nil {
		updates["LastName"] = *in.LastName
	}
	if in.PhoneNumber != nil {
		updates["PhoneNumber"] = in.PhoneNumber
	}

	// Update profile first if there are any changes beyond UpdatedAt
	if len(updates) > 1 {
		if err := s.accounts.UpdateProfile(ctx, uid, updates); err != nil {
			return nil, err
		}
	}

	// Update auth fields
	var displayName *string
	if in.FirstName != nil || in.LastName != nil {
		fn := ""
		ln := ""
		if in.FirstName != nil {
			fn = *in.FirstName
		}
		if in.LastName != nil {
			ln = *in.LastName
		}
		dn := strings.TrimSpace(strings.TrimSpace(fn + " " + ln))
		if dn != "" {
			displayName = &dn
		}
	}
	if in.Email != nil || in.PhoneNumber != nil || displayName != nil {
		if err := s.accounts.UpdateAuthUser(ctx, uid, in.Email, displayName, in.PhoneNumber); err != nil {
			return nil, err
		}
	}

	return s.accounts.GetProfile(ctx, uid)
}

func (s *UserService) DeleteUser(ctx context.Context, uid string) error {
	uid = strings.TrimSpace(uid)
	// Delete profile first (best-effort)
	_ = s.accounts.DeleteProfile(ctx, uid)
	// Then delete auth account
	return s.accounts.DeleteAuthUser(ctx, uid)
}

func (s *UserService) ForgotPassword(ctx context.Context, email string) (string, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return "", ErrValidation
	}
	return s.accounts.GeneratePasswordResetLink(ctx, email)
}

func (s *UserService) ChangePassword(ctx context.Context, uid string, newPassword string) error {
	uid = strings.TrimSpace(uid)
	newPassword = strings.TrimSpace(newPassword)
	if uid == "" || newPassword == "" {
		return ErrValidation
	}
	return s.accounts.UpdatePassword(ctx, uid, newPassword)
}

var (
	ErrValidation = &ValidationError{msg: "invalid input"}
)

type ValidationError struct{ msg string }

func (e *ValidationError) Error() string { return e.msg }
