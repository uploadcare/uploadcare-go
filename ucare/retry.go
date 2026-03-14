package ucare

// RetryConfig controls automatic retry of throttled (HTTP 429) requests.
// When nil in Config (the default), throttled requests fail immediately.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts.
	// 0 means no retries even if RetryConfig is set.
	MaxRetries int
	// MaxWaitSeconds limits retry waits, but the exact behavior depends on
	// which API returned HTTP 429:
	//   - REST API: if the server sends a positive Retry-After above this
	//     value, the request fails fast with ThrottleError instead of retrying.
	//     Fallback exponential backoff used when Retry-After is absent or
	//     invalid is not capped by this field.
	//   - Upload API: locally computed exponential backoff is clamped to this
	//     value because the upload API does not return Retry-After.
	// 0 means no cap.
	MaxWaitSeconds int
}

// expBackoff returns exponential backoff wait: 1, 2, 4, 8... capped at 30s.
func expBackoff(attempt int) int {
	wait := 1 << (attempt - 1)
	return min(wait, 30)
}
