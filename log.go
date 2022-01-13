package logging

import (
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strconv"
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
}

type logger struct {
	mu        sync.Mutex
	title     string
	separator string
	level     level
	flag      int
	w         io.Writer
}

func getCallInfo() callInfo {
	// skip 3 frames from stack to get right caller
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = sourceErr
		line = -1
	}
	return callInfo{file: file, line: line}
}

func (l *logger) isLevelHigherThanDefault(currentLevel level) bool {
	return sortedLevels[currentLevel] >= sortedLevels[l.level]
}

func (l *logger) getPrefix(level level, callInfo callInfo) (data []byte) {
	data = append(data, l.getDateTime()...)
	data = append(data, l.getTitle()...)
	data = append(data, l.getLevel(level)...)
	data = append(data, l.getCallerInfo(callInfo)...)
	return data
}

func (l *logger) getCallerInfo(callInfo callInfo) (data []byte) {
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

func (l *logger) getLevel(level level) (data []byte) {
	if l.flag&Labels != 0 {
		data = append(data, "LEVEL = "...)
	}
	data = append(data, fmt.Sprintf("[%s] %s ", level, l.separator)...)
	return data
}

func (l *logger) getTitle() (data []byte) {
	if l.title != "" {
		if l.flag&Labels != 0 {
			data = append(data, "TITLE = "...)
		}
		data = append(data, fmt.Sprintf("(%s) %s ", l.title, l.separator)...)
	}
	return data
}

func (l *logger) getDateTime() (data []byte) {
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

func (l *logger) log(level level, msg string) {
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
