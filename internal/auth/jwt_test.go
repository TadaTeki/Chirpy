package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "mysecret"
	expiresIn := time.Hour

	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	if tokenString == "" {
		t.Fatalf("MakeJWT returned empty token string")
	}

	parsedUserID, err := ValidateJWT(tokenString, tokenSecret)
	if err != nil {
		t.Fatalf("ValidateJWT returned error: %v", err)
	}

	if parsedUserID != userID {
		t.Errorf("ValidateJWT returned userID %v, want %v", parsedUserID, userID)
	}
}

func TestValidateJWT_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "mysecret"
	expiresIn := -time.Hour // 過去の時間を設定して期限切れにする

	tokenString, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	_, err = ValidateJWT(tokenString, tokenSecret)
	if err == nil {
		t.Fatalf("ValidateJWT did not return error for expired token")
	}

	if err != jwt.ErrTokenExpired {
		t.Errorf("ValidateJWT returned error %v, want jwt.ErrTokenExpired", err)
	}

}

func TestValidateJWT_InvalidToken(t *testing.T) {
	invalidTokenString := "invalid.token.string"
	tokenSecret := "mysecret"

	_, err := ValidateJWT(invalidTokenString, tokenSecret)
	if err == nil {
		t.Fatalf("ValidateJWT did not return error for invalid token")
	}
}
func TestValidateJWT_WrongSecret(t *testing.T) {
	userID := uuid.New()
	correctSecret := "correctsecret"
	wrongSecret := "wrongsecret"
	expiresIn := time.Hour

	tokenString, err := MakeJWT(userID, correctSecret, expiresIn)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}

	_, err = ValidateJWT(tokenString, wrongSecret)
	if err == nil {
		t.Fatalf("ValidateJWT did not return error for wrong secret")
	}
}
