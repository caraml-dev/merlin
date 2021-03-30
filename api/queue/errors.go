package queue

import "fmt"

type RetryableError struct {
	Message string
}

func (err RetryableError) Error() string {
	return fmt.Sprintf("got retryable error %v", err.Message)
}
