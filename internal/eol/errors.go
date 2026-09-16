package eol

import "fmt"

// NotFoundError is returned when the API responds with 404.
type NotFoundError struct {
	Kind string
	Name string
}

func (e *NotFoundError) Error() string {
	if e.Name == "" {
		return fmt.Sprintf("%s not found", e.Kind)
	}
	return fmt.Sprintf("%s %q not found", e.Kind, e.Name)
}

// RateLimitError is returned when the API responds with 429 after retries.
type RateLimitError struct {
	RetryAfter string
}

func (e *RateLimitError) Error() string {
	if e.RetryAfter == "" {
		return "API rate limit exceeded"
	}
	return fmt.Sprintf("API rate limit exceeded (retry after %s)", e.RetryAfter)
}
