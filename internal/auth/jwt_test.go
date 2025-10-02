package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWTAndValidateJWT(t *testing.T) {
	uuid1 := uuid.New()
	uuid2 := uuid.New()
	secret := "secret"
	duration := time.Hour
	t.Run("JWT Match", func(t *testing.T) {
		jwt, err := MakeJWT(uuid1, secret, duration)
		if err != nil {
			t.Fatal(err)
		}

		owner, err := ValidateJWT(jwt, secret)
		if err != nil {
			t.Fatal(err)
		}
		if owner != uuid1 {
			t.Errorf("expected owner %s, got %s", uuid1, owner)
		}
	})
	t.Run("JWT no Match", func(t *testing.T) {
		jwt, err := MakeJWT(uuid1, secret, duration)
		if err != nil {
			t.Fatal(err)
		}

		owner, err := ValidateJWT(jwt, secret)
		if err != nil {
			t.Fatal(err)
		}
		if owner == uuid2 {
			t.Errorf("expected owner %s, got %s", uuid2, uuid1)
		}
	})
}
