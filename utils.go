package gateway

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
