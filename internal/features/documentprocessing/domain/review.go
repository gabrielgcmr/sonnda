// internal/features/documentprocessing/domain/review.go
package domain

import "errors"

var ErrReviewConflict = errors.New("document is not a pending draft")
