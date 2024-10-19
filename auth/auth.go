package auth

import "github.com/google/uuid"

// generate_session_token generates a new session token using UUID.
func GenerateSessionToken() string {
	return uuid.New().String()
}
