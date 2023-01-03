package web_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/mroobert/tixer-pkgs/web"
)

func TestReadIDParam_ReturnsIdForValidValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want string
	}{
		{
			name: "Valid ID: abc123",
			want: "abc123",
		},
		{
			name: "Valid ID: abc123_xyz456",
			want: "abc123_xyz456",
		},
		{
			name: "Valid ID: user_12345",
			want: "user_12345",
		},
		{
			name: "Valid ID: customer_54321",
			want: "customer_54321",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, "http://api.com/tickets/", nil)
			if err != nil {
				t.Fatal(err)
			}
			params := httprouter.Params{
				httprouter.Param{
					Key:   "id",
					Value: tt.want,
				},
			}
			req = req.WithContext(context.WithValue(req.Context(), httprouter.ParamsKey, params))

			got, err := web.ReadIDParam(req)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("Expected id to be %s, but got %s", tt.want, got)
			}
		})
	}
}

func TestReadIDParam_ReturnsErrForInvalidValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   string
		want error
	}{
		{
			name: "Too long",
			id:   strings.Repeat("a", 1501),
			want: web.ErrIdTooLong,
		},
		{
			name: "Invalid UTF-8 characters",
			id:   "\xbd\xb2\x3d\xbc\x20\xe2\x8c\x98",
			want: web.ErrIdInvalidUTF8,
		},
		{
			name: "Prohibited character: /",
			id:   "abc/123",
			want: web.ErrIdProhibitedChar,
		},
		{
			name: "Prohibited character: \\",
			id:   "abc\\123",
			want: web.ErrIdProhibitedChar,
		},
		{
			name: "Prohibited pattern: . (single period)",
			id:   ".",
			want: web.ErrIdProhibitedPattern,
		},
		{
			name: "Prohibited pattern: .. (double period)",
			id:   "..",
			want: web.ErrIdProhibitedPattern,
		},
		{
			name: "Prohibited pattern: __ (starts with)",
			id:   "__abc123",
			want: web.ErrIdProhibitedPattern,
		},
		{
			name: "Prohibited pattern: __ (ends with)",
			id:   "abc123__",
			want: web.ErrIdProhibitedPattern,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, "http://api.com/tickets/", nil)
			if err != nil {
				t.Fatal(err)
			}
			params := httprouter.Params{
				httprouter.Param{
					Key:   "id",
					Value: tt.id,
				},
			}
			req = req.WithContext(context.WithValue(req.Context(), httprouter.ParamsKey, params))

			_, got := web.ReadIDParam(req)
			if !errors.Is(got, tt.want) {
				t.Errorf("Expected error to be %q, but got %q", got, tt.want)
			}
		})
	}
}
