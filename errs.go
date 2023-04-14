package logging

import (
	"fmt"
	"runtime"
	"strings"
)

type TraceableError interface {
	GetTrace() string
	GetAllStackFrames() []string
}

type traceableError struct {
	err   error
	frame string
}

func (trErr *traceableError) Error() string {
	if trErr.err != nil {
		return trErr.err.Error()
	}
	return "unknown (unspecified) error"
}

func (trErr *traceableError) Unwrap() error {
	return trErr.err
}

func (trErr *traceableError) GetTrace() string {
	return strings.Join(trErr.GetAllStackFrames(), "\n")
}

func (trErr *traceableError) GetAllStackFrames() (frames []string) {
	frames = append(frames, trErr.frame)
	if err, ok := trErr.err.(TraceableError); ok {
		frames = append(frames, err.GetAllStackFrames()...)
	} else {
		frames = append(frames, fmt.Sprintf("\terror occurred: %s", trErr.Error()))
	}

	return
}

func (trErr *traceableError) getCurrentStackFrame() string {
	if info := getCallInfo(); info.pc != 0 {
		if fn := runtime.FuncForPC(info.pc); fn != nil {
			return fmt.Sprintf("\t%s\n\t\t%s:%d", fn.Name(), info.file, info.line)
		}
		return fmt.Sprintf("\t%s:%d", info.file, info.line)
	}
	return "unknown stack frame"
}
