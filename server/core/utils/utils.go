package utils

import (
	"fmt"
	"strings"
)

func StringToArray[T any](str string) []T {
	strArr := strings.Split(str, ",")
	var arr []T
	for _, s := range strArr {
		var i T
		fmt.Sscanf(s, "%v", &i)
		arr = append(arr, i)
	}
	return arr
}

func Max[T ~int | ~int64 | ~int32 | ~float32 | ~float64 | ~uint | ~uint32 | ~uint64](x, y T) T {

	if x > y {
		return x
	} else {
		return y
	}
}

func Min[T ~int | ~int64 | ~int32 | ~float32 | ~float64 | ~uint | ~uint32 | ~uint64](x, y T) T {

	if x > y {
		return y
	} else {
		return x
	}
}
