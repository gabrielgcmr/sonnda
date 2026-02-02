// internal/kernel/observability/recovery.go
package observability

import (
	"runtime/debug"
)

type PanicEvent struct {
	Value any
	Stack string
}

func CapturePanic(fn func()) (event *PanicEvent) {
	defer func() {
		if r := recover(); r != nil {
			event = &PanicEvent{
				Value: r,
				Stack: string(debug.Stack()),
			}
		}
	}()
	fn()
	return nil
}
