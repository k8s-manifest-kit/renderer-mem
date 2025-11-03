package mem

import (
	"errors"
	"fmt"
)

var (
	// ErrObjectEmpty is returned when an object is empty or has nil internal data.
	ErrObjectEmpty = errors.New("object is empty or has nil internal data")
)

// sourceHolder wraps a Source with internal state for consistency with other renderers.
type sourceHolder struct {
	Source
}

// Validate checks if the Source configuration is valid.
func (h *sourceHolder) Validate() error {
	for i := range h.Objects {
		if len(h.Objects[i].Object) == 0 {
			return fmt.Errorf("%w at index %d", ErrObjectEmpty, i)
		}
	}

	return nil
}
