package logging

import (
	"fmt"
	"runtime"
	"strings"
)

type CodedError interface {
	error
	GetCode() int
	SetCode(int)
}

type TraceableError interface {
	CodedError
	GetTrace() string
	GetAllStackFrames() []string
}

type traceableError struct {
	err   error
	code  int
	frame string
}

func (trErr *traceableError) Error() string {
	if trErr.err != nil {
		return trErr.err.Error()
	}
	return "unknown (unspecified) error"
}

func (trErr *traceableError) GetCode() int {
	if trErr.code == 0 {
		if err, ok := trErr.err.(CodedError); ok {
			return err.GetCode()
		}
	}
	return trErr.code
}

func (trErr *traceableError) SetCode(code int) {
	trErr.code = code
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
