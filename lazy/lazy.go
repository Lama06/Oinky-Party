package lazy

type Value[T any] func() T

func New[T any](creator func() T) Value[T] {
	initialized := false
	var value T

	return func() T {
		if !initialized {
			value = creator()
			initialized = true
		}

		return value
	}
}
