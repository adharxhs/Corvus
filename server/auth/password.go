package auth

import (
	"crypto/rand"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argon2Memory  = 64 * 1024
	argon2Time    = 3
	argon2Threads = 2
	argon2KeyLen  = 32
	saltLen       = 16
)

// HashPassword hashes a password with Argon2id and returns an encoded string.
func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)
	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argon2Memory, argon2Time, argon2Threads,
		encodeB64(salt), encodeB64(hash),
	), nil
}

// VerifyPassword checks a plaintext password against an encoded hash.
func VerifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("malformed hash")
	}

	var memory, time, threads int
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads); err != nil {
		return false, fmt.Errorf("parsing hash params: %w", err)
	}

	salt, err := decodeB64(parts[4])
	if err != nil {
		return false, fmt.Errorf("decoding salt: %w", err)
	}
	expected, err := decodeB64(parts[5])
	if err != nil {
		return false, fmt.Errorf("decoding hash: %w", err)
	}

	actual := argon2.IDKey([]byte(password), salt, uint32(time), uint32(memory), uint8(threads), uint32(len(expected)))
	if len(actual) != len(expected) {
		return false, nil
	}
	for i := range actual {
		if actual[i] != expected[i] {
			return false, nil
		}
	}
	return true, nil
}
