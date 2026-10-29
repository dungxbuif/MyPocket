package usecase

type PasswordService interface {
	Hash(password string) (string, error)
	Check(password, hashedPassword string) bool
}
