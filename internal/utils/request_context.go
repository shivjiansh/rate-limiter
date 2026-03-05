package utils

import "github.com/google/uuid"

func EnsureRequestID(v string) string {
	if v != "" {
		return v
	}
	return uuid.NewString()
}
