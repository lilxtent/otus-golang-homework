package validators

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func NewValidationError(field string, err error) ValidationError {
	return ValidationError{
		Field: field,
		Err:   err,
	}
}
