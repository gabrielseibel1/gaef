package hasher

import "golang.org/x/crypto/bcrypt"

type PasswordHasherVerifier struct{}

// New TODO use function type instead of struct
func New() PasswordHasherVerifier {
	return PasswordHasherVerifier{}
}

func (phv PasswordHasherVerifier) GenerateFromPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (phv PasswordHasherVerifier) CompareHashAndPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
