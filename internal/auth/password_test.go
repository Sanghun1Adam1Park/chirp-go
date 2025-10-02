package auth

import "testing"

func TestHashPasswordAndCheckPassswordHash(t *testing.T) {
	cases := []struct {
		name       string
		password   string
		comparison string
		wantErr    bool
	}{
		{"Password Match", "mypassword1", "mypassword1", false},
		{"Password no Match", "mypassword1", "mypassword2", true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hashedPassword, err := HashPassword(tc.password)
			if err != nil {
				t.Fatalf("error hashing password: %v", err)
			}

			err = CheckPasswordHash(tc.comparison, hashedPassword)
			if tc.wantErr && err == nil {
				t.Errorf("error checking password hash: %v", err)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("error checking password hash: %v", err)
			}
		})
	}
}
