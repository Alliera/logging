package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
Configurable thread-safe custom Logger with method chaining.

Features:
- 5 levels of severity such as DEBUG, INFO, WARNING, ERROR and FATAL.
- configurable log format
- default level support to ignore all levels with lower priority

In example:

logger := logging.New(
		os.Stdout,
		"title",
		logging.Time | logging.ShortCaller,
		logging.WARNING,
		"|",
	)
*/

var sortedLevels = map[level]int{
	DEBUG:   0,
	INFO:    1,
	WARNING: 2,
	ERROR:   3,
	FATAL:   4,
}

type level string

type callInfo struct {
	file string
	line int
	pc   uintptr
}

type Logger struct {
	title     string
	separator string
	level     level
	flag      int
	w         io.Writer
}

var (
	exit   = os.Exit
	caller = runtime.Caller
)

func getCallInfo() callInfo {
	// skip 3 frames from stack to get right caller
	pc, file, line, ok := caller(3)
	if !ok {
		file = sourceErr
		line = -1
	}
	return callInfo{file: file, line: line, pc: pc}
}

func (l *Logger) isLevelHigherThanDefault(currentLevel level) bool {
	return sortedLevels[currentLevel] >= sortedLevels[l.level]
}

func (l *Logger) getPrefix(level level, callInfo callInfo) (data []byte) {
	data = append(data, l.getDateTime()...)
	data = append(data, l.getTitle()...)
	data = append(data, l.getLevel(level)...)
	data = append(data, l.getCallerInfo(callInfo)...)
	return data
}

func (l *Logger) getCallerInfo(callInfo callInfo) (data []byte) {
	if l.flag&(Caller|ShortCaller) != 0 {
		if l.flag&Labels != 0 {
			data = append(data, "SRC = "...)
		}
		if l.flag&ShortCaller != 0 {
			data = append(data, fmt.Sprintf("%s", filepath.Base(callInfo.file))...)
		} else {
			data = append(data, fmt.Sprintf("%s", callInfo.file)...)
		}
		data = append(data, ':')
		data = append(data, strconv.Itoa(callInfo.line)...)
		data = append(data, fmt.Sprintf(" %s ", l.separator)...)
	}
	return data
}

func (l *Logger) getLevel(level level) (data []byte) {
	if l.flag&Labels != 0 {
		data = append(data, "LEVEL = "...)
	}
	data = append(data, fmt.Sprintf("[%s] %s ", level, l.separator)...)
	return data
}

func (l *Logger) getTitle() (data []byte) {
	if l.title != "" {
		if l.flag&Labels != 0 {
			data = append(data, "TITLE = "...)
		}
		data = append(data, fmt.Sprintf("(%s) %s ", l.title, l.separator)...)
	}
	return data
}

func (l *Logger) getDateTime() (data []byte) {
	t := time.Now()
	if l.flag&Date != 0 {
		if l.flag&Labels != 0 {
			data = append(data, "DATE = "...)
		}
		data = append(data, t.Format("2006-01-02")...)
		data = append(data, fmt.Sprintf(" %s ", l.separator)...)
	}

	if l.flag&Time != 0 {
		if l.flag&Labels != 0 {
			data = append(data, "TIME =  "...)
		}
		data = append(data, t.Format("15:04:05")...)
		data = append(data, fmt.Sprintf(" %s ", l.separator)...)
	}
	return data
}

func (l *Logger) log(level level, msg string) {
	if !l.isLevelHigherThanDefault(level) {
		return
	}

	data := make([]byte, 0)
	data = append(l.getPrefix(level, getCallInfo()), data...)
	if l.flag&Labels != 0 {
		data = append(data, "MSG = "...)
	}

	if len(msg) == 0 {
		data = append(data, "Unknown error\n"...)
	} else {
		data = append(data, msg...)
		if msg[len(msg)-1] != '\n' {
			data = append(data, '\n')
		}
	}
	_, _ = l.w.Write(data)
}

func (l *Logger) getMsgFromError(err error, s []string) (msg string) {
	parts := append([]string{err.Error()}, s...)
	msg = strings.Join(parts, " "+l.separator+" ")
	if t, ok := err.(TraceableError); ok {
		msg = fmt.Sprintf("%s\n%s", msg, t.GetTrace())
	}
	return
}

type LoggerRegistry struct {
	loggers map[string]*Logger
	mu      sync.Mutex
}

func (r *LoggerRegistry) AddLogger(l *Logger) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.loggers[l.title]; ok {
		return fmt.Errorf("logger with name %s already exists", l.title)
	}
	r.loggers[l.title] = l
	return nil
}

func (r *LoggerRegistry) AddLoggerFromConfig(cfg Config) (*Logger, error) {
	l := NewFromConfig(cfg)
	if err := r.AddLogger(l); err != nil {
		return nil, err
	}
	return l, nil
}

func (r *LoggerRegistry) GetLogger(name string) (*Logger, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if l, ok := r.loggers[name]; ok {
		return l, nil
	}
	return nil, fmt.Errorf("logger with name %s does not exists", name)
}
