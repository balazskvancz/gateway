package gateway

import (
	"reflect"
	"testing"
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
