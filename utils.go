package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"time"
)

type filterFunc[T any] func(T) bool

// filter returns a new array from given
// array filtered by the given predicate function.
func filter[T any](arr []T, fn filterFunc[T]) []T {
	var (
		newArr  = make([]T, len(arr))
		counter = 0
	)

	for _, el := range arr {
		if fn(el) {
			newArr[counter] = el
			counter++
		}
	}
	return newArr[:counter]
}

type reduceFn[K, T any] func(K, T) K

// reduce is a generic function to implement the reduce function known from functional programming.
func reduce[T, K any](arr []T, fn reduceFn[K, T], initial K) K {
	var acc K = initial
	for _, el := range arr {
		acc = fn(acc, el)
	}
	return acc
}

// iuncludes a generic function to determine if one given array
// contains an element which is deepEqual to the predicate element.
func includes[T any](arr []T, el T) bool {
	for _, e := range arr {
		if reflect.DeepEqual(e, el) {
			return true
		}
	}
	return false
}

// createHash hashes and returns the given slice of bytes.
func createHash(plain []byte) []byte {
	ha := sha256.New()
	ha.Write([]byte(plain))

	s := ha.Sum(nil)

	dst := make([]byte, hex.EncodedLen(len(s)))
	hex.Encode(dst, s)

	return dst
}

const (
	day = time.Hour * 24
)

const (
	badTimesGiven = "the now time is before than since time"
)

func getElapsedTime(since, now time.Time) string {
	var timeString = ""

	elapsed := now.Sub(since)

	if elapsed.Milliseconds() < 0 {
		return badTimesGiven
	}

	if days := elapsed / day; days > 0 {
		timeString += fmt.Sprintf("%d days ", days)
		elapsed -= days * day
	}

	if hours := elapsed / time.Hour; hours > 0 {
		timeString += fmt.Sprintf("%d hours ", hours)
		elapsed -= hours * time.Hour
	}

	if minutes := elapsed / time.Minute; minutes > 0 {
		timeString += fmt.Sprintf("%d minutes ", minutes)
		elapsed -= minutes * time.Minute
	}

	timeString += (time.Duration(elapsed/time.Second) * time.Second).String()

	return timeString
}
