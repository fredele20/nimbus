package password

import "golang.org/x/crypto/bcrypt"

type passwordHasher struct {}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) error
}

func New() PasswordHasher {
	return &passwordHasher{}
}

func (p *passwordHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (p *passwordHasher) Verify(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
