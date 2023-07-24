package gateway

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"strings"
)

type filterFunc[T any] func(T) bool

func filter[T any](arr []T, fn filterFunc[T]) []T {
	newArr := make([]T, 0)
	for _, el := range arr {
		if fn(el) {
			newArr = append(newArr, el)
		}
	}
	return newArr
}

type reduceFn[K, T any] func(K, T) K

func reduce[T, K any](arr []T, fn reduceFn[K, T], initial K) K {
	var acc K = initial
	for _, el := range arr {
		acc = fn(acc, el)
	}
	return acc
}

func isNil[T any](val T) bool {
	return reflect.ValueOf(val).IsNil()
}

func includes[T any](arr []T, el T) bool {
	for _, e := range arr {
		if reflect.DeepEqual(e, el) {
			return true
		}
	}
	return false
}

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

func createHash(plain []byte) string {
	ha := sha256.New()
	ha.Write([]byte(plain))
	return hex.EncodeToString(ha.Sum(nil))
}
