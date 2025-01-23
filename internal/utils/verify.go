package utils

import (
	"math/rand"
)

// GenerateCode generates a random 4-digit code for user verification.
func GenerateCode() int {
	return rand.Intn(9000) + 1000
}
