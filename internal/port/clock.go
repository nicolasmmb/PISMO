package port

import "time"

// Clock allows injecting time source.
type Clock interface {
	Now() time.Time
}
