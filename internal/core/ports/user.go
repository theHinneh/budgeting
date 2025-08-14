package ports

import (
	"context"

	"github.com/theHinneh/budgeting/internal/core/models"
)

type CreateUserInput struct {
	UID         string
	Username    string
	Email       string
	FirstName   string
	LastName    string
	PhoneNumber *string
	Password    string // optional; if empty, UID must be provided and auth user must already exist
}

type UpdateUserInput struct {
	Username    *string
	Email       *string
	FirstName   *string
	LastName    *string
	PhoneNumber *string
}

type UserServicePort interface {
	CreateUser(ctx context.Context, in CreateUserInput) (string, error)
	GetUser(ctx context.Context, uid string) (*models.User, error)
	UpdateUser(ctx context.Context, uid string, in UpdateUserInput) (*models.User, error)
	DeleteUser(ctx context.Context, uid string) error
}

type UserAccountPort interface {
	CreateAuthUser(ctx context.Context, email, password, displayName string, phone *string) (string, error)
	GetAuthUser(ctx context.Context, uid string) error
	UpdateAuthUser(ctx context.Context, uid string, email *string, displayName *string, phone *string) error
	DeleteAuthUser(ctx context.Context, uid string) error

	SaveProfile(ctx context.Context, u *models.User) error
	GetProfile(ctx context.Context, uid string) (*models.User, error)
	UpdateProfile(ctx context.Context, uid string, updates map[string]interface{}) error
	DeleteProfile(ctx context.Context, uid string) error
}
