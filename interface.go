package logging

import (
	"io"
	"os"
	"strings"
)

func AddLogger(l *Logger) error {
	return registry.addLogger(l)
}

func AddLoggerFromConfig(cfg Config) (*Logger, error) {
	return registry.addLoggerFromConfig(cfg)
}

func GetLogger(name string) (*Logger, error) {
	return registry.getLogger(name)
}

func SetLevelForLogger(name string, level string) error {
	lvl, err := levelFromString(level)
	if err != nil {
		return err
	}
	return registry.setLevelForLogger(name, lvl)
}

func SetLevelForAll(level string) error {
	lvl, err := levelFromString(level)
	if err != nil {
		return err
	}
	registry.setLevelForAll(lvl)
	return nil
}

func ResetLevels() {
	registry.resetLevels()
}

func ResetLevel(loggerName string) error {
	l, err := registry.getLogger(loggerName)
	if err != nil {
		return err
	}
	l.resetLevel()
	return nil
}
func New(w io.Writer, title string, flag int, level level, separator string) *Logger {
	return &Logger{
		w:             w,
		title:         title,
		flag:          flag,
		level:         level,
		originalLevel: level,
		separator:     separator,
	}
}

func NewDefault(title string, l ...level) *Logger {
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

func NewFromConfig(cfg Config) *Logger {
	l := new(Logger)
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
	l.SetLevel(WARNING)
	l.originalLevel = WARNING
	lvl := level(strings.ToUpper(string(cfg.Level)))
	if lvl != "" {
		l.SetLevel(lvl)
		l.originalLevel = lvl
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

func (l *Logger) SetWriter(w io.Writer) *Logger {
	l.w = w
	return l
}

func (l *Logger) GetLevelInt() int {
	return sortedLevels[l.level]
}

func (l *Logger) SetFlags(flag int) *Logger {
	l.flag = l.flag | flag
	return l
}

func (l *Logger) UnsetFlags(flag int) *Logger {
	l.flag = l.flag &^ flag
	return l
}

func (l *Logger) SetSeparator(separator string) *Logger {
	l.separator = separator
	return l
}

func (l *Logger) SetLevel(level level) *Logger {
	l.level = level
	return l
}

func (l *Logger) GetWriter() io.Writer {
	return l.w
}

func (l *Logger) Info(msg string) {
	l.log(INFO, msg)
}

func (l *Logger) Warning(msg string) {
	l.log(WARNING, msg)
}

func (l *Logger) Debug(msg string) {
	l.log(DEBUG, msg)
}

func (l *Logger) Error(msg string) {
	l.log(ERROR, msg)
}

func (l *Logger) Fatal(msg string) {
	l.log(FATAL, msg)
	exit(1)
}

func (l *Logger) LogError(err error, s ...string) {
	if err == nil {
		return
	}
	l.log(ERROR, l.getMsgFromError(err, s))
}

func (l *Logger) LogFatal(err error, s ...string) {
	if err == nil {
		return
	}
	l.log(FATAL, l.getMsgFromError(err, s))
	exit(1)
}

func Trace(err error) TraceableError {
	if err == nil {
		return nil
	}
	trErr := new(traceableError)
	trErr.err = err
	trErr.frame = trErr.getCurrentStackFrame()
	return trErr
}
