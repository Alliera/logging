package logging

import (
	"io"
	"os"
)

func New(w io.Writer, title string, flag int, level level, separator string) *logger {
	return &logger{
		w:         w,
		title:     title,
		flag:      flag,
		level:     level,
		separator: separator,
	}
}

func NewDefault(title string) *logger {
	var lvl level
	if value, ok := os.LookupEnv("DEBUG"); ok && value == "1" {
		lvl = DEBUG
	} else {
		lvl = WARNING
	}
	return New(os.Stdout, title, ShortCaller, lvl, "--")
}

func (l *logger) SetWriter(w io.Writer) *logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.w = w
	return l
}

func (l *logger) SetFlags(flag int) *logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flag = l.flag | flag
	return l
}

func (l *logger) UnsetFlags(flag int) *logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flag = l.flag &^ flag
	return l
}

func (l *logger) SetSeparator(separator string) *logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.separator = separator
	return l
}

func (l *logger) SetLevel(level level) *logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
	return l
}

func (l *logger) Info(msg string) {
	l.log(INFO, msg)
}

func (l *logger) Warning(msg string) {
	l.log(WARNING, msg)
}

func (l *logger) Debug(msg string) {
	l.log(DEBUG, msg)
}

func (l *logger) Error(msg string) {
	l.log(ERROR, msg)
}

func (l *logger) Fatal(msg string) {
	l.log(FATAL, msg)
	os.Exit(1)
}
