package domain

import "golang.org/x/crypto/bcrypt"

type User struct {
	ID    string `json:"id" binding:"required"`
	Name  string `json:"name" bson:"name" binding:"required"`
	Email string `json:"email" bson:"email" binding:"required"`
}

type UserWithHashedPassword struct {
	ID             string `bson:"_id,omitempty"`
	Name           string `bson:"name"`
	Email          string `bson:"email"`
	HashedPassword string `bson:"password"`
}

func ToSimplifiedUser(u *UserWithHashedPassword) *User {
	return &User{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}
}

func FromSimplifiedUser(u *User, hashedPassword string) *UserWithHashedPassword {
	return &UserWithHashedPassword{
		ID:             u.ID,
		Name:           u.Name,
		Email:          u.Email,
		HashedPassword: hashedPassword,
	}
}

type PasswordHasherVerifier struct{}

func (phv PasswordHasherVerifier) GenerateFromPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func (phv PasswordHasherVerifier) CompareHashAndPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
