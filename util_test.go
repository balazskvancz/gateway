package gateway

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestFilter(t *testing.T) {
	type testCase[T any] struct {
		name     string
		input    []T
		fn       filterFunc[T]
		expected []T
	}

	ttForIntegers := []testCase[int]{
		{
			name:  "nothing to filter",
			input: []int{},
			fn: func(val int) bool {
				return true
			},
			expected: []int{},
		},
		{
			name:  "filter func always returns true",
			input: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			fn: func(val int) bool {
				return true
			},
			expected: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			name:  "filter func always returns false",
			input: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			fn: func(val int) bool {
				return false
			},
			expected: []int{},
		},
		{
			name:  "filter func filters only the even numbers",
			input: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			fn: func(val int) bool {
				return val%2 == 0
			},
			expected: []int{2, 4, 6, 8, 10},
		},
	}

	for _, tc := range ttForIntegers {
		t.Run(tc.name, func(t *testing.T) {
			got := filter(tc.input, tc.fn)

			var (
				lExpected = len(tc.expected)
				lGot      = len(got)
			)

			if lExpected != lGot {
				t.Errorf("expected length of filtered array: %d; got lenght: %v\n", lExpected, lGot)
			}

			if !reflect.DeepEqual(tc.expected, got) {
				t.Errorf("expected and got arrays are not deep equal")
			}

		})
	}
}

func TestReduce(t *testing.T) {
	type testCase[T, K any] struct {
		name         string
		input        []T
		fn           reduceFn[T, K]
		initialValue K
		expected     K
	}

	ttForIntegers := []testCase[int, int]{
		{
			name:  "empty array returns the initial value",
			input: []int{},
			fn: func(i1, i2 int) int {
				return i1
			},
			initialValue: 10,
			expected:     10,
		},
		{
			name:  "not empty array and sums the values",
			input: []int{1, 2, 3, 4, 5},
			fn: func(acc, curr int) int {
				return acc + curr
			},
			initialValue: 0,
			expected:     15,
		},

		{
			name:  "not empty array only sums the odd values",
			input: []int{1, 2, 3, 4, 5},
			fn: func(acc, curr int) int {
				if curr%2 == 0 {
					return acc
				}

				return acc + curr
			},
			initialValue: 0,
			expected:     9,
		},
	}

	for _, tc := range ttForIntegers {
		t.Run(tc.name, func(t *testing.T) {
			got := reduce(tc.input, tc.fn, tc.initialValue)

			if got != tc.expected {
				t.Errorf("expected value: %d; got: %d\n", tc.expected, got)
			}
		})
	}

	// Some more complex tests.
	t.Run("integers into arrays", func(t *testing.T) {
		var (
			inputArr = []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

			expectedArr = []int{2, 4, 6, 8, 10}
		)

		var r reduceFn[[]int, int] = func(acc []int, curr int) []int {
			if curr%2 != 0 {
				return acc
			}

			return append(acc, curr)
		}

		got := reduce(inputArr, r, []int{})

		if !reflect.DeepEqual(got, expectedArr) {
			t.Errorf("int arrays are not deep equal")
		}
	})

	t.Run("struct mutation", func(t *testing.T) {
		type tComplex struct {
			val string
		}

		strs := []string{"foo", "foobarbazfoo", "bar", "barfoo", "barbaz", ""}

		var r reduceFn[*tComplex, string] = func(acc *tComplex, curr string) *tComplex {
			if len(curr) > len(acc.val) {
				acc.val = curr
			}
			return acc
		}

		got := reduce(strs, r, &tComplex{})

		const expected = "foobarbazfoo"

		if got.val != expected {
			t.Errorf("expected: %s; got: %s\n", expected, got.val)
		}
	})
}

func TestIncludes(t *testing.T) {
	type testCase struct {
		name             string
		arr              []string
		searchEl         string
		expectedIncludes bool
	}

	tt := []testCase{
		{
			name:             "the fn returns true if the given string is included",
			arr:              []string{"foo", "bar", "baz"},
			searchEl:         "foo",
			expectedIncludes: true,
		},
		{
			name:             "the fn returns false if the given string is not included",
			arr:              []string{"foo", "bar", "baz"},
			searchEl:         "fo",
			expectedIncludes: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := includes(tc.arr, tc.searchEl)

			if got != tc.expectedIncludes {
				t.Errorf("expected: %v; got: %v\n", tc.expectedIncludes, got)
			}
		})
	}
}

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
			parts := getUrlParts(tc.url)

			if len(parts) != len(tc.expected) {
				t.Fatalf("expected length: %d, got length: %d", len(tc.expected), len(parts))
			}

			if !reflect.DeepEqual(tc.expected, parts) {
				t.Errorf("got slice is not as expected")
			}
		})
	}
}

type getContextFn func(t *testing.T) context.Context

func TestGetValueFromContext(t *testing.T) {
	var defaultKey contextKey = "value-key"

	type testCase struct {
		name        string
		getCtx      getContextFn
		expected    string
		expectedErr error
	}

	tt := []testCase{
		{
			name: "the function returns error if the given ctx is nil",
			getCtx: func(t *testing.T) context.Context {
				return nil
			},
			expected:    "",
			expectedErr: errContextIsNil,
		},
		{
			name: "the function returns error if cant parse it the value",
			getCtx: func(t *testing.T) context.Context {
				ctx := context.WithValue(context.Background(), defaultKey, 5)
				return ctx
			},
			expected:    "",
			expectedErr: errKeyInContextIsNotPresent,
		},
		{
			name: "the function returns default value for the type if the key is not present",
			getCtx: func(t *testing.T) context.Context {
				type testCtxKey string

				var key testCtxKey = "foo"

				return context.WithValue(context.Background(), key, "bar")
			},
			expected:    "",
			expectedErr: errKeyInContextIsNotPresent,
		},
		{
			name: "the function returns the stored value associated with the key",
			getCtx: func(t *testing.T) context.Context {
				return context.WithValue(context.Background(), defaultKey, "bar")
			},
			expected:    "bar",
			expectedErr: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx := tc.getCtx(t)

			val, err := getValueFromContext[string](ctx, defaultKey)

			if val != tc.expected {
				t.Errorf("expected value: %s; got value: %s\n", tc.expected, val)
			}

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error: %v; got error: %v\n", tc.expectedErr, err)
			}
		})
	}
}

func TestGetElapsedTime(t *testing.T) {
	type testCase struct {
		name     string
		since    time.Time
		now      time.Time
		expected string
	}

	tt := []testCase{
		{
			name:     "function returns error string if the now time is before since time",
			since:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			now:      time.Date(2023, 1, 1, 7, 0, 0, 0, time.UTC),
			expected: badTimesGiven,
		},

		{
			name:     "function returns `0s` if there is no elapsed time",
			since:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			now:      time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: "0s",
		},
		{
			name:     "function returns the proper elapsed time (only seconds)",
			since:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			now:      time.Date(2023, 1, 1, 12, 0, 15, 0, time.UTC),
			expected: "15s",
		},
		{
			name:     "function returns the proper elapsed time (minutes and seconds)",
			since:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			now:      time.Date(2023, 1, 1, 12, 5, 24, 0, time.UTC),
			expected: "5 minutes 24s",
		},
		{
			name:     "function returns the proper elapsed time (hours, minutes and seconds)",
			since:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			now:      time.Date(2023, 1, 1, 14, 5, 24, 0, time.UTC),
			expected: "2 hours 5 minutes 24s",
		},
		{
			name:     "function returns the proper elapsed time (days, hours, minutes and seconds)",
			since:    time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			now:      time.Date(2023, 1, 23, 14, 5, 24, 0, time.UTC),
			expected: "22 days 2 hours 5 minutes 24s",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got := getElapsedTime(tc.since, tc.now)

			if got != tc.expected {
				t.Errorf("expected result: %s; got result: %s\n", tc.expected, got)
			}
		})
	}
}

func TestCreateHash(t *testing.T) {
	type testCase struct {
		input    []byte
		expected []byte
	}

	var tt = []testCase{
		{
			input:    []byte("mock-test-1"),
			expected: []byte("1613ace6d16db2ec3ccd55a85d4125f3a83b5315b961b14fe6e50951d9551b54"),
		},
		{
			input:    []byte("api-gateway-2022"),
			expected: []byte("e31cb50b87d55d14e6c18561075cba2cc22cdd0a8ce10d1dd167a3adbd08224c"),
		},
		{
			input:    []byte("1"),
			expected: []byte("6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b"),
		},
	}

	for _, tc := range tt {
		if output := createHash(tc.input); !reflect.DeepEqual(output, tc.expected) {
			t.Errorf("expected: %s got: %s\n", string(tc.expected), string(output))
		}
	}
}
