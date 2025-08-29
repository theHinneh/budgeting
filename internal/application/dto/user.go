package dto

type CreateUserInput struct {
	UID         string
	Username    string
	Email       string
	FirstName   string
	LastName    string
	PhoneNumber *string
}

type UpdateUserInput struct {
	Username    *string
	Email       *string
	FirstName   *string
	LastName    *string
	PhoneNumber *string
}
