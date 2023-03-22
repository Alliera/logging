package logging

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"strings"
	"testing"
	"time"
)

type WriterMock struct {
	mock.Mock
}

func (w *WriterMock) Write(p []byte) (n int, err error) {
	w.Called(p)
	return 1, nil
}

func TestNewFromConfig_Default(t *testing.T) {
	cfg := Config{
		Title: "test",
	}
	l := NewFromConfig(cfg)

	assert.Equal(t, "test", l.title)
	assert.Equal(t, "--", l.separator)
	assert.Equal(t, WARNING, l.level)
	assert.Equal(t, os.Stdout, l.w)
	assert.Equal(t, 0, l.flag)
}

func TestNewFromConfig(t *testing.T) {
	cfg := Config{
		Title:             "test",
		Separator:         "||",
		Level:             ERROR,
		Direction:         "stderr",
		EnableDate:        true,
		EnableTime:        true,
		EnableCaller:      true,
		EnableShortCaller: true,
		EnableLabels:      true,
	}

	l := NewFromConfig(cfg)
	assert.Equal(t, cfg.Title, l.title)
	assert.Equal(t, cfg.Separator, l.separator)
	assert.Equal(t, cfg.Level, l.level)
	assert.Equal(t, os.Stderr, l.w)
	assert.Equal(t, Date|Time|Caller|Labels|ShortCaller, l.flag)

	cfg.Direction = "stdout"
	l = NewFromConfig(cfg)
	assert.Equal(t, os.Stdout, l.w)
}

func TestNewFromConfig_CustomFileOutput(t *testing.T) {
	defer func() {
		_ = os.Remove("clerk_test.log")
	}()

	cfg := Config{
		Title:     "test",
		Direction: "clerk_test.log",
	}
	l := NewFromConfig(cfg)

	assert.Equal(t, "clerk_test.log", l.w.(*os.File).Name())
	fileStat, _ := os.Stat("clerk_test.log")
	assert.Equal(t, "-rw-r--r--", fileStat.Mode().String())
	assert.Equal(t, "-rw-r--r--", fileStat.Mode().Perm().String())
	assert.Equal(t, int64(0), fileStat.Size())

	l.Warning("some warning msg")
	l.Error("some error msg")
	LogContent, _ := os.ReadFile("clerk_test.log")
	assert.Equal(t,
		"(test) -- [WARNING] -- some warning msg\n"+
			"(test) -- [ERROR] -- some error msg\n",
		string(LogContent))
}

func TestResetLevel(t *testing.T) {
	registry.clear()
	loggerName := "logger"
	wrongName := "wrongLogger"
	l := NewDefault(loggerName, ERROR)
	_ = AddLogger(l)
	l.SetLevel(DEBUG)
	assert.Equal(t, l.level, DEBUG)
	err := ResetLevel(loggerName)
	assert.Equal(t, err, nil)
	assert.Equal(t, l.level, ERROR)
	err = ResetLevel(wrongName)
	assert.Equal(t, err.Error(), fmt.Sprintf("logger with name %s does not exists", wrongName))
}

func TestAddLogger(t *testing.T) {
	registry.clear()
	loggerName := "logger.test"
	l := NewDefault(loggerName, ERROR)
	err := AddLogger(l)
	assert.Equal(t, err, nil)
	err = AddLogger(l)
	assert.Equal(t, err.Error(), fmt.Sprintf("logger with name %s already exists", loggerName))
}

func TestAddLoggerFromConfig(t *testing.T) {
	registry.clear()
	title := "test"
	cfg := Config{
		Title: title,
	}
	logger, err := AddLoggerFromConfig(cfg)
	assert.Equal(t, err, nil)
	assert.Equal(t, title, logger.title)
	_, err = AddLoggerFromConfig(cfg)
	assert.Equal(t, err.Error(), fmt.Sprintf("logger with name %s already exists", title))
}

func TestGetLogger(t *testing.T) {
	registry.clear()
	title := "test"
	notExistsTitle := "testNotExists"
	l := NewDefault(title, ERROR)
	_ = AddLogger(l)
	logger, err := GetLogger(title)
	assert.Equal(t, err, nil)
	assert.Equal(t, logger.title, title)
	_, err = GetLogger(notExistsTitle)
	assert.Equal(t, err.Error(), fmt.Sprintf("logger with name %s does not exists", notExistsTitle))
}

func TestSetLevelForLogger(t *testing.T) {
	registry.clear()
	title := "logger"
	notExistsTitle := "testNotExists"
	wrongLevel := "wrongLevel"
	_, _ = AddLoggerFromConfig(Config{
		Title: title,
		Level: DEBUG,
	})
	err := SetLevelForLogger(title, wrongLevel)
	assert.Equal(t, err.Error(), fmt.Sprintf("level %s invalid", strings.ToUpper(wrongLevel)))
	err = SetLevelForLogger(notExistsTitle, "iNfO")
	assert.Equal(t, err.Error(), fmt.Sprintf("logger with name %s does not exists", notExistsTitle))
	err = SetLevelForLogger(title, "iNfO")
	assert.Equal(t, err, nil)
	logger, _ := GetLogger(title)
	assert.Equal(t, logger.level, INFO)
}

func TestResetLevels(t *testing.T) {
	registry.clear()
	l1 := NewDefault("l1", ERROR)
	_ = AddLogger(l1)
	l2 := NewDefault("l2", INFO)
	_ = AddLogger(l2)
	l1.SetLevel(WARNING)
	l2.SetLevel(WARNING)
	assert.Equal(t, l1.level, WARNING)
	assert.Equal(t, l2.level, WARNING)
	ResetLevels()
	assert.Equal(t, l1.level, ERROR)
	assert.Equal(t, l2.level, INFO)
}

func TestSetLevelForAll(t *testing.T) {
	registry.clear()
	l1 := NewDefault("l1", ERROR)
	_ = AddLogger(l1)
	l2 := NewDefault("l2", INFO)
	_ = AddLogger(l2)
	err := SetLevelForAll("WArNING")
	assert.Equal(t, err, nil)
	assert.Equal(t, l1.level, WARNING)
	assert.Equal(t, l2.level, WARNING)
	err = SetLevelForAll("inValid")
	assert.Equal(t, err.Error(), fmt.Sprintf("level %s invalid", "INVALID"))
}

func TestNewDefault(t *testing.T) {
	var l *Logger

	l = NewDefault("")
	assert.Equal(t, os.Stdout, l.w)
	assert.Equal(t, WARNING, l.level)

	t.Setenv("DEBUG", "1")
	l = NewDefault("")
	assert.Equal(t, DEBUG, l.level)
	l = NewDefault("", ERROR)
	assert.Equal(t, DEBUG, l.level)

	t.Setenv("DEBUG", "")
	l = NewDefault("")
	assert.Equal(t, WARNING, l.level)
	l = NewDefault("", FATAL)
	assert.Equal(t, FATAL, l.level)
}

func TestSetSeparator(t *testing.T) {
	w := WriterMock{}
	l := Logger{level: DEBUG, w: &w}
	msg := "asdasd"

	w.On("Write", []byte(fmt.Sprintf("[WARNING]  %s\n", msg)))
	l.Warning(msg)

	l.SetSeparator(DefaultSeparator)
	w.On("Write", []byte(fmt.Sprintf("[WARNING] -- %s\n", msg)))
	l.Warning(msg)

	l.SetSeparator("|")
	w.On("Write", []byte(fmt.Sprintf("[WARNING] | %s\n", msg)))
	l.Warning(msg)

	w.AssertNumberOfCalls(t, "Write", 3)
	w.AssertExpectations(t)
}

func TestGetCallerInfo(t *testing.T) {
	l := New(&WriterMock{}, "test", Labels|Caller, DEBUG, DefaultSeparator)
	ci := callInfo{
		file: "/dev/null",
		line: 42,
		pc:   0,
	}
	i := string(l.getCallerInfo(ci))
	assert.Equal(t, "SRC = /dev/null:42 -- ", i)
	l.UnsetFlags(Caller)
	l.SetFlags(ShortCaller)
	i = string(l.getCallerInfo(ci))
	assert.Equal(t, "SRC = null:42 -- ", i)
}

func TestSetFlags(t *testing.T) {
	l := Logger{}
	l.SetFlags(Date | Date)
	assert.Equal(t, Date, l.flag)
	l.SetFlags(Date | Time)
	assert.Equal(t, Date|Time, l.flag)
	l.SetFlags(ShortCaller)
	assert.Equal(t, Date|Time|ShortCaller, l.flag)
}

func TestUnsetFlags(t *testing.T) {
	l := Logger{}
	l.SetFlags(Date | Time | Labels | Caller | ShortCaller)
	l.UnsetFlags(Time)
	assert.Equal(t, Date|Labels|Caller|ShortCaller, l.flag)
	l.UnsetFlags(Labels | Caller)
	assert.Equal(t, Date|ShortCaller, l.flag)
	l.UnsetFlags(Labels)
	assert.Equal(t, Date|ShortCaller, l.flag)
}

func TestIsLevelHigherThanDefault(t *testing.T) {
	l := Logger{}
	l.level = FATAL
	assert.False(t, l.isLevelHigherThanDefault(DEBUG))
	l.level = DEBUG
	assert.True(t, l.isLevelHigherThanDefault(DEBUG))
	l.level = DEBUG
	assert.True(t, l.isLevelHigherThanDefault(FATAL))
}

func TestLogNoFlags(t *testing.T) {
	w := WriterMock{}
	msg := "this is info message"
	w.On("Write", []byte(fmt.Sprintf("(title)  [INFO]  %s\n", msg)))
	l := Logger{title: "title", w: &w}
	l.log(INFO, msg)
	w.AssertExpectations(t)
}

func TestLogEmptyNoFlags(t *testing.T) {
	w := WriterMock{}
	msg := ""
	w.On("Write", []byte("(title)  [INFO]  Unknown error\n"))
	l := Logger{title: "title", w: &w}
	l.log(INFO, msg)
	w.AssertExpectations(t)
}

func TestLogWithFlags(t *testing.T) {
	w := WriterMock{}
	l := Logger{title: "title", w: &w}
	l.SetFlags(Date | Time | Labels)

	msg := "this is info message"
	now := time.Now()

	w.On("Write", []byte(fmt.Sprintf(
		"DATE = %v  TIME =  %v  TITLE = (title)  "+
			"LEVEL = [INFO]  MSG = %s\n",
		now.Format("2006-01-02"),
		now.Format("15:04:05"),
		msg,
	)))

	l.Info(msg)
	w.AssertExpectations(t)
}

func TestDifferentLogLevels(t *testing.T) {
	w := WriterMock{}

	w.On("Write", []byte("[WARNING]  msg\n"))
	w.On("Write", []byte("[DEBUG]  msg\n"))
	w.On("Write", []byte("[ERROR]  msg\n"))
	w.On("Write", []byte("[INFO]  msg\n"))

	l := Logger{level: WARNING, w: &w}
	l.Warning("msg")
	l.Debug("msg")
	l.Error("msg")
	l.SetLevel(ERROR)
	l.Error("msg")
	l.SetWriter(&w)
	l.Debug("msg")
	l.Info("msg")
	l.SetLevel(INFO)
	l.Info("msg")
	assert.Equal(t, &w, l.GetWriter())

	w.AssertNumberOfCalls(t, "Write", 4)
}

func TestTrace(t *testing.T) {
	assert.Equal(t, nil, Trace(nil))

	msg := "asdasdasd"
	err := errors.New(msg)

	tracedErr := Trace(err)
	assert.NotEqual(t, nil, tracedErr)
	assert.Equal(t, msg, tracedErr.Error())

	traceableErr, ok := tracedErr.(TraceableError)
	assert.True(t, ok)
	assert.Equal(t, 2, len(traceableErr.GetAllStackFrames()))

	tracedTracedErr := Trace(tracedErr)
	assert.NotEqual(t, nil, tracedTracedErr)
	assert.Equal(t, msg, tracedTracedErr.Error())

	traceableErr, ok = tracedTracedErr.(TraceableError)
	assert.True(t, ok)
	assert.Equal(t, 3, len(traceableErr.GetAllStackFrames()))
}

func TestLogError(t *testing.T) {
	l := NewDefault("test").UnsetFlags(ShortCaller)
	w := &WriterMock{}
	l.SetWriter(w)

	l.LogError(nil)

	w.On("Write", []byte("(test) -- [ERROR] -- some error\n"))
	w.On("Write", []byte("(test) -- [ERROR] -- some error -- some info\n"))
	w.On("Write", []byte("(test) -- [ERROR] -- some error -- some info -- more info\n"))

	err := errors.New("some error")
	l.LogError(err)
	l.LogError(err, "some info")
	l.LogError(err, "some info", "more info")

	w.AssertNumberOfCalls(t, "Write", 3)
}

func TestFatal(t *testing.T) {
	l := NewDefault("test").UnsetFlags(ShortCaller)
	w := &WriterMock{}
	l.SetWriter(w)

	exit = func(i int) {}
	w.On("Write", []byte("(test) -- [FATAL] -- some error\n"))
	l.Fatal("some error")
	w.AssertExpectations(t)
	w.AssertNumberOfCalls(t, "Write", 1)
}

func TestLogFatal(t *testing.T) {
	l := NewDefault("test").UnsetFlags(ShortCaller)
	w := &WriterMock{}
	l.SetWriter(w)

	exit = func(i int) {}
	l.LogFatal(nil)
	w.On("Write", []byte("(test) -- [FATAL] -- some error\n"))
	err := errors.New("some error")
	l.LogFatal(err)
	w.AssertExpectations(t)
	w.AssertNumberOfCalls(t, "Write", 1)
}

func TestLogErrorTraceableError(t *testing.T) {
	l := NewDefault("test").UnsetFlags(ShortCaller)
	w := &WriterMock{}
	l.SetWriter(w)

	w.On("Write", []byte("(test) -- [ERROR] -- some error\nfake frame\n\terror occurred: some error\n"))
	w.On("Write", []byte("(test) -- [ERROR] -- some error -- some info\nfake frame\n\terror occurred: some error\n"))
	w.On("Write", []byte("(test) -- [ERROR] -- some error -- some info -- more info\nfake frame\n\terror occurred: some error\n"))
	w.On("Write", []byte("(test) -- [ERROR] -- unknown (unspecified) error\nfake frame\n\terror occurred: unknown (unspecified) error\n"))

	var err *traceableError

	err = &traceableError{
		err:   errors.New("some error"),
		frame: "fake frame",
	}

	l.LogError(err)
	l.LogError(err, "some info")
	l.LogError(err, "some info", "more info")

	err = &traceableError{frame: "fake frame"}
	l.LogError(err)

	w.AssertNumberOfCalls(t, "Write", 4)
}

func TestGetCallInfo(t *testing.T) {
	caller = func(i int) (pc uintptr, file string, line int, ok bool) {
		return 0, "/dev/null", 42, true
	}
	ci := getCallInfo()
	assert.Equal(t, callInfo{file: "/dev/null", line: 42, pc: 0}, ci)

	caller = func(i int) (pc uintptr, file string, line int, ok bool) { return }
	ci = getCallInfo()
	assert.Equal(t, callInfo{file: sourceErr, line: -1, pc: 0}, ci)
}

func TestGetCurrentStackFrame(t *testing.T) {
	var (
		frame string
		err   traceableError
	)

	caller = func(i int) (pc uintptr, file string, line int, ok bool) {
		return 0, "/dev/null", 42, true
	}
	frame = err.getCurrentStackFrame()
	assert.Equal(t, "unknown stack frame", frame)

	caller = func(i int) (pc uintptr, file string, line int, ok bool) {
		return 1, "/dev/null", 42, true
	}
	frame = err.getCurrentStackFrame()
	assert.Equal(t, "\t/dev/null:42", frame)
}
