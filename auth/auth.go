package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
)

const (
	memory      = 64 * 1024
	iterations  = 3
	parallelism = 2
	saltLength  = 16
	keyLength   = 32
)

func GenerateSessionToken() string {
	return uuid.New().String()
}

func HashPassword(plainPassword string) (hashedPassword string, err error) {

	salt := make([]byte, saltLength)
	// fmt.Printf("salt1: %v\n", salt)
	rand.Read(salt)
	// fmt.Printf("salt2: %v\n", salt)
	hashBytes := argon2.IDKey([]byte(plainPassword), salt, iterations, memory, parallelism, keyLength)
	// fmt.Printf("hashBytes: %v\n)", hashBytes)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hashBytes)
	hashedPassword = fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, memory, iterations, parallelism, b64Salt, b64Hash)
	// fmt.Printf("hashedPassword: %v\n", hashedPassword)
	return hashedPassword, nil

}
