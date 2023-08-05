package gateway

import (
	"testing"
	"time"
)

func TestGetHealthCheckInterval(t *testing.T) {
	type testCase struct {
		name     string
		input    string
		expected time.Duration
	}

	tt := []testCase{
		{
			name:     "the functions returns 0 duration if the input is empty",
			input:    "",
			expected: 0,
		},
		{
			name:     "the functions returns 0 duration if the unit is not correct",
			input:    "10h",
			expected: 0,
		},
		{
			name:     "the functions returns 0 duration if the value is not parseable",
			input:    "mock1s",
			expected: 0,
		},
		{
			name:     "the functions returns 5 seconds duration if correct input",
			input:    "5s",
			expected: 5 * time.Second,
		},
		{
			name:     "the functions returns 10 minutes duration if correct input",
			input:    "10m",
			expected: 10 * time.Minute,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := getHealthCheckInterval(tc.input)

			if got != tc.expected {
				t.Errorf("expected value: %d; got value: %d\n", tc.expected, got)
			}
		})
	}
}
