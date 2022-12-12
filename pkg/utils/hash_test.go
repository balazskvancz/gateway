package utils

import "testing"

func TestHash(t *testing.T) {

	var tt = []struct {
		input    string
		expected string
	}{
		{
			input:    "mock-test-1",
			expected: "1613ace6d16db2ec3ccd55a85d4125f3a83b5315b961b14fe6e50951d9551b54",
		},
		{
			input:    "api-gateway-2022",
			expected: "e31cb50b87d55d14e6c18561075cba2cc22cdd0a8ce10d1dd167a3adbd08224c",
		},
		{
			input:    "1",
			expected: "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b"},
	}

	for _, tc := range tt {
		if output := CreateHash(tc.input); output != tc.expected {
			t.Errorf("expected: %s; got: %s\n", output, tc.expected)
		}
	}
}
