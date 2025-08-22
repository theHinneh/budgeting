package application

var (
	ErrValidation = &ValidationError{msg: "invalid input"}
)

type ValidationError struct{ msg string }

func (e *ValidationError) Error() string { return e.msg }
