// internal/kernel/persistence/errors.go
package persistence

import "errors"

// ErrPersistenceFailure identifies a persistence failure independently of its driver.
// Implementations wrap or join it with the original cause for internal diagnostics.
var ErrPersistenceFailure = errors.New("persistence failure")
