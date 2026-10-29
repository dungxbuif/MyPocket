package auth

import "golang.org/x/crypto/bcrypt"

type Password struct{}

func NewPasswordService() *Password {
	return &Password{}
}

func (p *Password) Hash(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (p *Password) Check(password, hashedPassword string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
