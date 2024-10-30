package util

func Ptr[T any](input T) *T {
	return &input
}
