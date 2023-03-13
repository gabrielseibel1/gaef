package service

import (
	"context"
	"errors"

	"github.com/gabrielseibel1/gaef-user-service/domain"
)

// dependencies

type PasswordHasher interface {
	GenerateFromPassword(password string) (string, error)
}
type PasswordVerifier interface {
	CompareHashAndPassword(hashedPassword, password string) error
}
type Creator interface {
	Create(user *domain.UserWithHashedPassword, ctx context.Context) (string, error)
}
type ByIDReader interface {
	ReadByID(id string, ctx context.Context) (*domain.User, error)
}
type ByEmailReader interface {
	ReadSensitiveByEmail(email string, ctx context.Context) (*domain.UserWithHashedPassword, error)
}
type Updater interface {
	Update(user *domain.User, ctx context.Context) (*domain.User, error)
}
type Deleter interface {
	Delete(id string, ctx context.Context) error
}

// implementation

type Service struct {
	passwordHasher   PasswordHasher
	passwordVerifier PasswordVerifier
	creator          Creator
	byIDReader       ByIDReader
	byEmailReader    ByEmailReader
	updater          Updater
	deleter          Deleter
}

func New(passwordHasher PasswordHasher, passwordVerifier PasswordVerifier, creator Creator, byIDReader ByIDReader, byEmailReader ByEmailReader, updater Updater, deleter Deleter) *Service {
	return &Service{
		passwordHasher:   passwordHasher,
		passwordVerifier: passwordVerifier,
		creator:          creator,
		byIDReader:       byIDReader,
		byEmailReader:    byEmailReader,
		updater:          updater,
		deleter:          deleter,
	}
}

func (s Service) Create(user *domain.User, password string, ctx context.Context) (string, error) {
	_, err := s.byEmailReader.ReadSensitiveByEmail(user.Email, ctx)
	if err == nil {
		return "", errors.New("email is taken")
	}
	hash, err := s.passwordHasher.GenerateFromPassword(password)
	if err != nil {
		return "", err
	}
	return s.creator.Create(domain.FromSimplifiedUser(user, hash), ctx)
}

func (ss Service) Login(email, password string, ctx context.Context) (*domain.User, error) {
	u, err := ss.byEmailReader.ReadSensitiveByEmail(email, ctx)
	if err != nil {
		return nil, err
	}
	err = ss.passwordVerifier.CompareHashAndPassword(u.HashedPassword, password)
	if err != nil {
		return nil, err
	}
	return domain.ToSimplifiedUser(u), nil
}

func (ss Service) Read(id string, ctx context.Context) (*domain.User, error) {
	user, err := ss.byIDReader.ReadByID(id, ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ss Service) Update(user *domain.User, ctx context.Context) (*domain.User, error) {
	user, err := ss.updater.Update(user, ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ss Service) Delete(id string, ctx context.Context) error {
	return ss.deleter.Delete(id, ctx)
}
