package gateway

import "reflect"

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
