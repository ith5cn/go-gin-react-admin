package middleware

import "testing"

func TestIsOperLogMethod(t *testing.T) {
	cases := []struct {
		method string
		want   bool
	}{
		{"GET", false},
		{"HEAD", false},
		{"OPTIONS", false},
		{"POST", true},
		{"PUT", true},
		{"PATCH", true},
		{"DELETE", true},
	}
	for _, tc := range cases {
		if got := isOperLogMethod(tc.method); got != tc.want {
			t.Fatalf("isOperLogMethod(%q) = %v, want %v", tc.method, got, tc.want)
		}
	}
}

func TestSanitizeOperLogData(t *testing.T) {
	input := `{"username":"admin","password":"secret","accessToken":"abc","Authorization":"Bearer xyz","name":"kept"}`
	want := `{"username":"admin","password":"***","accessToken":"***","Authorization":"***","name":"kept"}`
	if got := sanitizeOperLogData(input); got != want {
		t.Fatalf("sanitizeOperLogData() = %q, want %q", got, want)
	}
}
