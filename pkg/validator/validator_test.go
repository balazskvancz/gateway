package validator

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/balazskvancz/gateway/pkg/gcontext"
	"github.com/balazskvancz/gateway/pkg/utils"
)

func TestNew(t *testing.T) {
	tt := []struct {
		name  string
		field string
		key   string
		hash  string

		expected *Validator
	}{
		{
			name:     "the functions returns a nil pointer, if the field is empty",
			field:    "",
			key:      "",
			expected: nil,
		},
		{
			name:     "the functions returns a nil pointer, if the key is empty",
			field:    "test-field",
			key:      "",
			expected: nil,
		},
		{
			name:  "the functions returns a pointer",
			field: "test-field",
			key:   "test-key",
			hash:  "62af8704764faf8ea82fc61ce9c4c3908b6cb97d463a634e9e587d7c885db0ef",
			expected: &Validator{
				FieldName: "test-field",
				SecretKey: "test-key",
				HashedKey: "62af8704764faf8ea82fc61ce9c4c3908b6cb97d463a634e9e587d7c885db0ef",
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := New(tc.field, tc.key)

			if tc.expected == nil {
				if got != nil {
					t.Errorf("expected nil pointer, but got a pointer")
				}

				return
			}

			if got.FieldName != tc.field {
				t.Errorf("expected field: %s; got: %s\n", tc.field, got.FieldName)
			}

			if got.SecretKey != tc.key {
				t.Errorf("expected key: %s; got: %s\n", tc.key, got.SecretKey)
			}

			if got.HashedKey != tc.hash {
				t.Errorf("expected hash: %s; got: %s\n", tc.hash, got.HashedKey)
			}

		})
	}
}

func TestValidateHeader(t *testing.T) {
	// tc1: No signature field.
	// tc2: Empty signature.
	// tc3: Not empty, but not valid signature.
	// tc4: Not empty and valid signature.

	secretKey := "api-gw-2022-thesis"
	fieldKey := "signature"

	emptyHeader := http.Header{}
	emptyHeader.Add(fieldKey, "")

	badHeader := http.Header{}
	b64Bad := utils.EncodeB64("not_good_value")
	badHeader.Add(fieldKey, b64Bad)

	b64Good := utils.EncodeB64(utils.CreateHash(secretKey))
	goodHeader := http.Header{}
	goodHeader.Add(fieldKey, b64Good)

	tt := []struct {
		name     string
		header   http.Header
		expected bool
	}{
		{
			name:     "the functions returns false, if there is no key",
			header:   http.Header{},
			expected: false,
		},
		{
			name:     "the functions returns false, if the key is empty",
			header:   emptyHeader,
			expected: false,
		},
		{
			name:     "the functions returns false, if the key is not valid",
			header:   badHeader,
			expected: false,
		},
		{
			name:     "the functions returns true, if the key is valid",
			header:   goodHeader,
			expected: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api", nil)
			req.Header = tc.header

			ctx := gcontext.New(nil, req)

			validator := New(fieldKey, secretKey)

			if validator == nil {
				t.Fatalf("got nil validator\n")
			}

			got := validator.ValidateHeader(ctx)

			if tc.expected != got {
				t.Errorf("expected to be: %v; got: %v\n", tc.expected, got)
			}

		})
	}
}
