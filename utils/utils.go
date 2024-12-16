package utils

func Abs[T int16](n T) T {
	if n <= 0 {
		return -n
	}
	return n
}

func Min[T uint16](a, b T) T {
	if a < b {
		return a
	}
	return b
}