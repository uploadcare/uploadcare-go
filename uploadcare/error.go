package uploadcare

import "fmt"

type ThrottleErr struct {
	RetryAfter int
}

func (e ThrottleErr) Error() string {
	return fmt.Sprintf(
		"Request was throttled. Expected available in %d second",
		e.RetryAfter,
	)
}
