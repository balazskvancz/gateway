package gateway

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
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

// isNil returns whether a value is nil or not.
func isNil[T any](val T) bool {
	return reflect.ValueOf(val).IsNil()
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

// getUrlParts returns the given URL splitted into a slice of strings
// where the delimiter is the / char. It also normalizeses
// the url which means it adds a / prefix if it is not present
// and also strips the query params if there any.
func getUrlParts(url string) []string {
	strSlash := string(slash)
	if !strings.HasPrefix(url, strSlash) {
		url = strSlash + url
	}

	queryStart := strings.IndexRune(url, query)
	if queryStart > 0 {
		url = url[:queryStart]
	}

	return strings.Split(url, strSlash)[1:]
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

// getValueFromContext returns the value by the given from the the given context.
func getValueFromContext[T any](ctx context.Context, key contextKey) (T, error) {
	var def T

	if ctx == nil {
		return def, errContextIsNil
	}

	val, ok := ctx.Value(key).(T)
	if !ok {
		return def, errKeyInContextIsNotPresent
	}

	return val, nil
}

const (
	day = time.Hour * 24
)

func getElapsedTime(t time.Time) string {
	var timeString = ""

	elapsed := time.Since(t)

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
