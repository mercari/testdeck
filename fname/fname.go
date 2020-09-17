package fname

import (
	"regexp"
	"runtime"
)

/*
fname.go: A fname function that returns the name of the current function. Currently this is only used in unit tests but may be useful in other cases as well
*/

// optionalLevel is the level of nesting (may be specified if you are nesting fname in other utility methods)
func Fname(optionalLevel ...int) string {
	level := 0
	if len(optionalLevel) > 0 {
		level = optionalLevel[0]
	}
	pc := make([]uintptr, 1)
	runtime.Callers(2+level, pc)
	fs := runtime.CallersFrames(pc)
	frame, _ := fs.Next()

	re := regexp.MustCompile(`\w+$`)
	fname := []byte(frame.Func.Name())
	return string(re.Find(fname))
}
