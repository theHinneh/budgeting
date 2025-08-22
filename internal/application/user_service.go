package application

import (
	"context"
	"strings"
	"time"

	"github.com/theHinneh/budgeting/internal/application/ports"
	"github.com/theHinneh/budgeting/internal/domain"
)

type UserService struct {
	userRepo      ports.UserRepository
	authenticator ports.UserAuthenticator
}

func NewUserService(userRepo ports.UserRepository, authenticator ports.UserAuthenticator) *UserService {
	return &UserService{userRepo: userRepo, authenticator: authenticator}
}

var _ ports.UserServicePort = (*UserService)(nil)

func (s *UserService) CreateUser(ctx context.Context, in ports.CreateUserInput) (string, error) {
	// Basic validation
	if strings.TrimSpace(in.Username) == "" || strings.TrimSpace(in.Email) == "" || strings.TrimSpace(in.FirstName) == "" || strings.TrimSpace(in.LastName) == "" {
		return "", ErrValidation
	}

	var uid string
	password := strings.TrimSpace(in.Password)
	var err error
	if password != "" {
		// Create auth user
		display := strings.TrimSpace(strings.TrimSpace(in.FirstName + " " + in.LastName))
		u, authErr := s.authenticator.CreateAuthUser(ctx, in.Email, password, display, in.PhoneNumber)
		if authErr != nil {
			return "", authErr
		}
		uid = u
	} else {
		uid = strings.TrimSpace(in.UID)
		if uid == "" {
			return "", ErrValidation
		}
		// Verify auth user exists
		if authErr := s.authenticator.GetAuthUser(ctx, uid); authErr != nil {
			return "", authErr
		}
	}

	// Build and store profile
	user := domain.NewUser(uid, in.Username, in.Email, in.FirstName, in.LastName, in.PhoneNumber)

	_, err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return "", err
	}
	return uid, nil
}

func (s *UserService) GetUser(ctx context.Context, uid string) (*domain.User, error) {
	return s.userRepo.GetUser(ctx, strings.TrimSpace(uid))
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

	user, err := s.userRepo.GetUser(ctx, uid)
	if err != nil {
		return nil, err
	}

	if in.Username != nil {
		user.Username = *in.Username
	}
	if in.Email != nil {
		user.Email = *in.Email
	}
	if in.FirstName != nil {
		user.FirstName = *in.FirstName
	}
	if in.LastName != nil {
		user.LastName = *in.LastName
	}
	if in.PhoneNumber != nil {
		user.PhoneNumber = in.PhoneNumber
	}
	user.UpdatedAt = time.Now().UTC()

	updatedUser, err := s.userRepo.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}

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
		if err := s.authenticator.UpdateAuthUser(ctx, uid, in.Email, displayName, in.PhoneNumber); err != nil {
			return updatedUser, err
		}
	}

	return updatedUser, nil
}

func (s *UserService) DeleteUser(ctx context.Context, uid string) error {
	uid = strings.TrimSpace(uid)
	_ = s.userRepo.DeleteUser(ctx, uid)
	return s.authenticator.DeleteAuthUser(ctx, uid)
}

func (s *UserService) ForgotPassword(ctx context.Context, email string) (string, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return "", ErrValidation
	}
	return s.authenticator.GeneratePasswordResetLink(ctx, email)
}

func (s *UserService) ChangePassword(ctx context.Context, uid string, newPassword string) error {
	uid = strings.TrimSpace(uid)
	newPassword = strings.TrimSpace(newPassword)
	if uid == "" || newPassword == "" {
		return ErrValidation
	}
	return s.authenticator.UpdatePassword(ctx, uid, newPassword)
}
