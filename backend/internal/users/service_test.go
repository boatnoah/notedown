package users_test

import (
	"context"
	"testing"

	"github.com/boatnoah/notedown/internal/storage/memory"
	"github.com/boatnoah/notedown/internal/users"
	"github.com/boatnoah/notedown/pkg/types"
)

func newService() *users.Service {
	return users.NewService(memory.NewUserRepository())
}

func validInput() users.RegisterInput {
	return users.RegisterInput{
		Name:     "Alice",
		Email:    "alice@example.com",
		Username: "alice",
		Password: "securepassword",
		Pfp:      types.PfpBlue,
	}
}

func TestRegister_Success(t *testing.T) {
	svc := newService()
	user, err := svc.Register(context.Background(), validInput())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID == "" {
		t.Error("expected non-empty ID")
	}
	if user.Email != "alice@example.com" {
		t.Errorf("got email %q, want alice@example.com", user.Email)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	svc := newService()
	in := validInput()
	in.Email = "not-an-email"
	_, err := svc.Register(context.Background(), in)
	if err != users.ErrInvalidEmail {
		t.Errorf("got %v, want ErrInvalidEmail", err)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	svc := newService()
	in := validInput()
	in.Password = "short"
	_, err := svc.Register(context.Background(), in)
	if err != users.ErrWeakPassword {
		t.Errorf("got %v, want ErrWeakPassword", err)
	}
}

func TestRegister_InvalidPfp(t *testing.T) {
	svc := newService()
	in := validInput()
	in.Pfp = types.PfpPreset("neon-pink")
	_, err := svc.Register(context.Background(), in)
	if err != users.ErrInvalidPfp {
		t.Errorf("got %v, want ErrInvalidPfp", err)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	svc := newService()
	if _, err := svc.Register(context.Background(), validInput()); err != nil {
		t.Fatal(err)
	}
	in := validInput()
	in.Username = "alice2"
	_, err := svc.Register(context.Background(), in)
	if err != users.ErrDuplicateEmail {
		t.Errorf("got %v, want ErrDuplicateEmail", err)
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	svc := newService()
	if _, err := svc.Register(context.Background(), validInput()); err != nil {
		t.Fatal(err)
	}
	in := validInput()
	in.Email = "other@example.com"
	_, err := svc.Register(context.Background(), in)
	if err != users.ErrDuplicateUsername {
		t.Errorf("got %v, want ErrDuplicateUsername", err)
	}
}

func TestRegister_MissingFields(t *testing.T) {
	svc := newService()
	_, err := svc.Register(context.Background(), users.RegisterInput{})
	if err != users.ErrMissingFields {
		t.Errorf("got %v, want ErrMissingFields", err)
	}
}

func TestRegister_DefaultPfp(t *testing.T) {
	svc := newService()
	in := validInput()
	in.Pfp = ""
	user, err := svc.Register(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Pfp != types.PfpBlue {
		t.Errorf("got pfp %q, want %q", user.Pfp, types.PfpBlue)
	}
}

func TestRegister_DisplayNameEmail(t *testing.T) {
	svc := newService()
	in := validInput()
	in.Email = "Alice <alice@example.com>"
	user, err := svc.Register(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "alice@example.com" {
		t.Errorf("got email %q, want bare alice@example.com", user.Email)
	}
}
