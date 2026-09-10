package download

import (
	"errors"

	"audivo-media-downloader/internal/models"
)

type UserError struct {
	Value *models.UserError
}

func (e *UserError) Error() string {
	if e == nil || e.Value == nil {
		return "unknown Audivo error"
	}
	return e.Value.Message
}

func NewError(code, message string, retryable bool) error {
	return &UserError{Value: &models.UserError{Code: code, Message: message, Retryable: retryable}}
}

func NewDetailedError(code, message, details string, retryable bool) error {
	return &UserError{Value: &models.UserError{Code: code, Message: message, Details: details, Retryable: retryable}}
}

func AsUserError(err error) *models.UserError {
	if err == nil {
		return nil
	}
	var userErr *UserError
	if errors.As(err, &userErr) && userErr != nil && userErr.Value != nil {
		return userErr.Value
	}
	return &models.UserError{Code: "operation_failed", Message: "Não foi possível concluir a operação.", Details: err.Error(), Retryable: true}
}
