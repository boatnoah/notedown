package users

import (
	"context"
	"errors"
	"net/mail"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"

	"github.com/boatnoah/notedown/pkg/types"
)

var (
	ErrInvalidEmail  = errors.New("invalid email address")
	ErrWeakPassword  = errors.New("password must be at least 8 characters")
	ErrInvalidPfp    = errors.New("pfp must be one of: blue, green, red, yellow, purple, orange")
	ErrMissingFields = errors.New("name, email, username, and password are required")
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

type RegisterInput struct {
	Name     string
	Email    string
	Username string
	Password string
	Pfp      types.PfpPreset
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (*types.User, error) {
	if in.Name == "" || in.Email == "" || in.Username == "" || in.Password == "" {
		return nil, ErrMissingFields
	}

	addr, err := mail.ParseAddress(in.Email)
	if err != nil {
		return nil, ErrInvalidEmail
	}
	email := addr.Address // strip any display-name wrapping

	if utf8.RuneCountInString(in.Password) < 8 {
		return nil, ErrWeakPassword
	}

	if in.Pfp == "" {
		in.Pfp = types.PfpBlue
	}
	if !in.Pfp.Valid() {
		return nil, ErrInvalidPfp
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &types.User{
		Name:     in.Name,
		Email:    email,
		Username: in.Username,
		Pfp:      in.Pfp,
	}

	if err := s.repo.Create(ctx, user, string(hash)); err != nil {
		return nil, err
	}

	return user, nil
}
