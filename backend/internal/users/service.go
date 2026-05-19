package users

import (
	"context"
	"errors"
	"net/mail"

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
	if _, err := mail.ParseAddress(in.Email); err != nil {
		return nil, ErrInvalidEmail
	}
	if len(in.Password) < 8 {
		return nil, ErrWeakPassword
	}
	if !in.Pfp.Valid() {
		return nil, ErrInvalidPfp
	}

	exists, err := s.repo.ExistsByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateEmail
	}

	exists, err = s.repo.ExistsByUsername(ctx, in.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrDuplicateUsername
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &types.User{
		Name:     in.Name,
		Email:    in.Email,
		Username: in.Username,
		Pfp:      in.Pfp,
	}

	if err := s.repo.Create(ctx, user, string(hash)); err != nil {
		return nil, err
	}

	return user, nil
}
