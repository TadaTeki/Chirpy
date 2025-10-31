package auth

import (
	"testing"

	"github.com/alexedwards/argon2id"
)

func TestHashPassword(t *testing.T) {
	password := "mypassword"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash == "" {
		t.Fatalf("HashPassword returned emtpy hash")
	}

	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		t.Fatalf("ComparePasswordAndHash returned error: %v", err)
	}
	if !match {
		t.Error("Hashpassword: hash does not match original password")
	}
}

func TestHashPassword_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"normal password", "secret123", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Hashpassword() error = %v, wantErr %v", err, tt.wantErr)
			}

			if hash == "" {
				t.Fatalf("HashPassword() returned emtpy hash")
			}

			match, err := argon2id.ComparePasswordAndHash(tt.password, hash)
			if err != nil {
				t.Fatalf("ComparePasswordAndHash() error = %v", err)
			}
			if !match {
				t.Errorf("HashPassword() produced a hash that doesn't match")
			}
		})
	}
}

func TestCheckPasswordHash_TableDriven(t *testing.T) {

	validPassword := "validPass123"
	validHash, err := argon2id.CreateHash(validPassword, argon2id.DefaultParams)
	if err != nil {
		t.Fatalf("faild to create hash for test setup: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
		wantErr  bool
	}{
		{
			name:     "correct password",
			password: validPassword,
			hash:     validHash,
			want:     true,
			wantErr:  false,
		},
		{
			name:     "incorrect password",
			password: "wrong password",
			hash:     validHash,
			want:     false,
			wantErr:  false,
		},
		{
			name:     "malformed hash",
			password: validPassword,
			hash:     "not_a_real_hash",
			want:     false,
			wantErr:  true,
		},
		{
			name:     "emtpy password",
			password: "",
			hash:     validHash,
			want:     false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckPasswordHash(tt.password, tt.hash)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CheckPasswordHash() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("CheckPasswordHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
