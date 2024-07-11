package response

import "runtime"

func ErrorLine() int {
	_, _, line, ok := runtime.Caller(1)
	if !ok {
		line = 0
	}
	return line
}
