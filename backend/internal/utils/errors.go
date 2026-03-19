package utils

import (
	"errors"
	"net/http"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type AppError struct {
	StatusCode int
	Code       string
	Message    string
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(statusCode int, code, message string, err error) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Err:        err,
	}
}

func BadRequestError(message string) *AppError {
	return NewAppError(http.StatusBadRequest, "BAD_REQUEST", message, nil)
}

func UnauthorizedError(message string) *AppError {
	return NewAppError(http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func ForbiddenError(message string) *AppError {
	return NewAppError(http.StatusForbidden, "FORBIDDEN", message, nil)
}

func NotFoundError(message string) *AppError {
	return NewAppError(http.StatusNotFound, "NOT_FOUND", message, nil)
}

func ConflictError(message string) *AppError {
	return NewAppError(http.StatusConflict, "CONFLICT", message, nil)
}

func InternalError(err error) *AppError {
	return NewAppError(http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", err)
}

func NormalizeError(err error) *AppError {
	if err == nil {
		return InternalError(nil)
	}

	appError, ok := err.(*AppError)
	if ok {
		return appError
	}

	return InternalError(err)
}

// IsDuplicateKeyError checks if the error is a duplicate key violation.
// Supports both raw MySQL driver error (1062) and GORM's translated ErrDuplicatedKey.
func IsDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062
	}
	return false
}
