package utils

func Abs[T int16](n T) T {
	if n <= 0 {
		return -n
	}
	return n
}