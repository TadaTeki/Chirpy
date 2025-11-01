package auth

import (
	"strings"
	"testing"
)

func TestGetBearerToken(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string][]string
		wantToken  string
		wantErr    bool
		wantErrStr string
	}{
		{
			name: "wrong_fmt",
			headers: map[string][]string{
				"fruits": []string{"apple", "secret123"},
			},
			wantToken:  "",
			wantErr:    true,
			wantErrStr: "no authorization header found",
		},
		{
			name: "correct_fmt_w_bad_authorization",
			headers: map[string][]string{
				"Authorization": []string{"BearerTokenWithoutspace"},
				"fruits":        []string{"apple", "secret123"},
			},
			wantToken:  "",
			wantErr:    true,
			wantErrStr: "no beearer in the header",
		},
		{
			name: "correct_fmt_w_correct_authorization",
			headers: map[string][]string{
				"Authorization": []string{"Bearer TokenWith space"},
				"fruits":        []string{"apple", "secret123"},
			},
			wantToken:  "TokenWith space",
			wantErr:    false,
			wantErrStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GetBearerToken(tt.headers)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if tt.wantErrStr != "" && !strings.Contains(err.Error(), tt.wantErrStr) {
					t.Fatalf("expected error %q, got %v", tt.wantErrStr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if token != tt.wantToken {
				t.Errorf("expected token %q, got %q", tt.wantToken, token)
			}
		})
	}

}
