package stringutils

func IsEmpty(str string) bool {
	return ""==str
}

func IsNotEmpty(str string) bool {
	return !IsEmpty(str)
}
