// internal/domain/repository/errors.go
package repository

import "errors"

// ErrRepositoryFailure identifies a persistence failure independently of its driver.
// Implementations wrap or join it with the original cause for internal diagnostics.
var ErrRepositoryFailure = errors.New("repository failure")
