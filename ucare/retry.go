package ucare

type RetryConfig struct {
	MaxRetries     int
	MaxWaitSeconds int
}

func expBackoff(attempt int) int {
	wait := 1 << (attempt - 1)
	return min(wait, 30)
}
