package utils

import "testing"

func TestGetUrlParts(t *testing.T) {
	tt := []struct {
		name     string
		url      string
		expected []string
	}{
		{
			name:     "the function returns the parts, when normal url called",
			url:      "/foo/bar/baz",
			expected: []string{"foo", "bar", "baz"},
		},
		{
			name:     "the function returns the parts, when / prefix missing",
			url:      "foo/bar/baz",
			expected: []string{"foo", "bar", "baz"},
		},
		{
			name:     "the function returns the parts, when there are query params",
			url:      "foo/bar/baz?foo=yes",
			expected: []string{"foo", "bar", "baz"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			parts := GetUrlParts(tc.url)

			if len(parts) != len(tc.expected) {
				t.Fatalf("expected length: %d, got length: %d", len(tc.expected), len(parts))
			}

			for i, p := range tc.expected {
				if p != parts[i] {
					t.Errorf("expected part: %s, got part: %s\n", p, parts[i])
				}

			}

		})

	}

}
