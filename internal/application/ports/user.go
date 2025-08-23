package ports

import (
	"context"

	"github.com/theHinneh/budgeting/internal/domain"
)

type CreateUserInput struct {
	UID         string
	Username    string
	Email       string
	FirstName   string
	LastName    string
	PhoneNumber *string
	Password    string
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
	GetUser(ctx context.Context, uid string) (*domain.User, error)
	UpdateUser(ctx context.Context, uid string, in UpdateUserInput) (*domain.User, error)
	DeleteUser(ctx context.Context, uid string) error

	ForgotPassword(ctx context.Context, email string) (string, error)
	ChangePassword(ctx context.Context, uid string, newPassword string) error
}

type UserAccountPort interface {
	CreateAuthUser(ctx context.Context, email, password, displayName string, phone *string) (string, error)
	GetAuthUser(ctx context.Context, uid string) error
	UpdateAuthUser(ctx context.Context, uid string, email *string, displayName *string, phone *string) error
	DeleteAuthUser(ctx context.Context, uid string) error

	UpdatePassword(ctx context.Context, uid string, newPassword string) error
	GeneratePasswordResetLink(ctx context.Context, email string) (string, error)

	SaveProfile(ctx context.Context, u *domain.User) error
	GetProfile(ctx context.Context, uid string) (*domain.User, error)
	UpdateProfile(ctx context.Context, uid string, updates map[string]interface{}) error
	DeleteProfile(ctx context.Context, uid string) error
}

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUser(ctx context.Context, uid string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	DeleteUser(ctx context.Context, uid string) error
	ListAllUserIDs(ctx context.Context) ([]string, error)
}

type UserAuthenticator interface {
	CreateAuthUser(ctx context.Context, email, password, displayName string, phone *string) (string, error)
	GetAuthUser(ctx context.Context, uid string) error
	UpdateAuthUser(ctx context.Context, uid string, email *string, displayName *string, phone *string) error
	DeleteAuthUser(ctx context.Context, uid string) error
	UpdatePassword(ctx context.Context, uid string, newPassword string) error
	GeneratePasswordResetLink(ctx context.Context, email string) (string, error)
}
