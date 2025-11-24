package security

import "golang.org/x/crypto/bcrypt"

// HashPassword hashes a plaintext password with the provided cost, defaulting to bcrypt.DefaultCost when cost <= 0.
func HashPassword(plaintext string, cost int) (string, error) {
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), cost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ComparePassword compares a hashed password with a plaintext candidate.
func ComparePassword(hash, plaintext string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plaintext))
}
