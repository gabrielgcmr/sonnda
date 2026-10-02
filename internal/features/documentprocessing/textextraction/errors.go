// internal/features/documentprocessing/textextraction/errors.go
package textextraction

import "errors"

var ErrUnreadablePDF = errors.New("PDF has no usable selectable text")
