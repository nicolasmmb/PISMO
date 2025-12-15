package clock

import "time"

// SystemClock provides time from the system.
type SystemClock struct{}

// Now returns the current time.
func (SystemClock) Now() time.Time {
	return time.Now()
}
