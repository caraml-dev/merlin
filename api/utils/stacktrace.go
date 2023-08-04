package utils

import (
	"bytes"
	"fmt"
	"runtime"
)

func StackTrace(additionalCallerSkip int) string {
	// Ask runtime.Callers for up to 5 PCs, excluding current function and runtime.Callers itself,
	// in addition to excluding the `additionalCallerSkip` number of frames.
	// So, skip=(2+additionalCallerSkip).
	pc := make([]uintptr, 5)
	n := runtime.Callers(2+additionalCallerSkip, pc)
	if n == 0 {
		return ""
	}
	pc = pc[:n] // pass only valid pcs to runtime.CallersFrames
	frames := runtime.CallersFrames(pc)

	buf := bytes.NewBufferString("")
	for {
		frame, more := frames.Next()

		// Process this frame.
		fmt.Fprintf(buf, "%s\n", frame.Function)
		fmt.Fprintf(buf, "\t%s:%d\n", frame.File, frame.Line)

		// Check whether there are more frames to process after this one.
		if !more {
			break
		}
	}

	return buf.String()
}
