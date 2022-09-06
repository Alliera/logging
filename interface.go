package logging

import (
	"io"
	"os"
	"strings"
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

func NewDefault(title string, l ...level) *logger {
	var lvl level
	if value, ok := os.LookupEnv("DEBUG"); ok && value == "1" {
		lvl = DEBUG
	} else if len(l) > 0 {
		lvl = l[0]
	} else {
		lvl = WARNING
	}
	return New(os.Stdout, title, ShortCaller, lvl, DefaultSeparator)
}

type Config struct {
	Title             string `yaml:"title"`
	Separator         string `yaml:"separator"`
	Level             level  `yaml:"level"`
	Direction         string `yaml:"direction"`
	EnableDate        bool   `yaml:"enable_date"`
	EnableTime        bool   `yaml:"enable_time"`
	EnableLabels      bool   `yaml:"enable_labels"`
	EnableCaller      bool   `yaml:"enable_caller"`
	EnableShortCaller bool   `yaml:"enable_short_caller"`
}

func NewFromConfig(cfg Config) *logger {
	l := new(logger)
	l.title = cfg.Title

	if cfg.Direction == "stdout" {
		l.SetWriter(os.Stdout)
	} else if cfg.Direction == "stderr" {
		l.SetWriter(os.Stderr)
	} else if cfg.Direction == "" {
		l.SetWriter(os.Stdout)
	} else {
		f, _ := os.OpenFile(cfg.Direction, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		l.SetWriter(f)
	}
	if cfg.Level == "" {
		l.SetLevel(WARNING)
	} else {
		l.SetLevel(level(strings.ToUpper(string(cfg.Level))))
	}

	if cfg.Separator == "" {
		l.SetSeparator("--")
	} else {
		l.SetSeparator(cfg.Separator)
	}

	if cfg.EnableDate {
		l.SetFlags(Date)
	}
	if cfg.EnableTime {
		l.SetFlags(Time)
	}
	if cfg.EnableCaller {
		l.SetFlags(Caller)
	}
	if cfg.EnableShortCaller {
		l.SetFlags(ShortCaller)
	}
	if cfg.EnableLabels {
		l.SetFlags(Labels)
	}

	return l
}

func (l *logger) SetWriter(w io.Writer) *logger {
	l.w = w
	return l
}

func (l *logger) SetFlags(flag int) *logger {
	l.flag = l.flag | flag
	return l
}

func (l *logger) UnsetFlags(flag int) *logger {
	l.flag = l.flag &^ flag
	return l
}

func (l *logger) SetSeparator(separator string) *logger {
	l.separator = separator
	return l
}

func (l *logger) SetLevel(level level) *logger {
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
	exit(1)
}

func (l *logger) LogError(err error, s ...string) {
	if err == nil {
		return
	}
	l.log(ERROR, l.getMsgFromError(err, s))
}

func (l *logger) LogFatal(err error, s ...string) {
	if err == nil {
		return
	}
	l.log(FATAL, l.getMsgFromError(err, s))
	exit(1)
}

func Trace(err error) error {
	if err == nil {
		return nil
	}
	trErr := new(traceableError)
	trErr.err = err
	trErr.frame = trErr.getCurrentStackFrame()
	return trErr
}
