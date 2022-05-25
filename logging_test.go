package logging

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
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

func TestNewDefault(t *testing.T) {
	l := NewDefault("")

	if l.w != os.Stdout {
		t.Errorf("Output direction must be Stdout")
	}
	if l.level != WARNING && l.level != DEBUG {
		t.Errorf("Log level must be WARNING or DEBUG, has: %v", l.level)
	}
}

func TestSetFlags(t *testing.T) {
	l := logger{}
	l.SetFlags(Date | Date)
	assert.Equal(t, Date, l.flag)
	l.SetFlags(Date | Time)
	assert.Equal(t, Date|Time, l.flag)
	l.SetFlags(ShortCaller)
	assert.Equal(t, Date|Time|ShortCaller, l.flag)
}

func TestUnsetFlags(t *testing.T) {
	l := logger{}
	l.SetFlags(Date | Time | Labels | Caller | ShortCaller)
	l.UnsetFlags(Time)
	assert.Equal(t, Date|Labels|Caller|ShortCaller, l.flag)
	l.UnsetFlags(Labels | Caller)
	assert.Equal(t, Date|ShortCaller, l.flag)
	l.UnsetFlags(Labels)
	assert.Equal(t, Date|ShortCaller, l.flag)
}

func TestIsLevelHigherThanDefault(t *testing.T) {
	l := logger{}
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
	l := logger{title: "title", w: &w}
	l.log(INFO, msg)
	w.AssertExpectations(t)
}

func TestLogWithFlags(t *testing.T) {
	w := WriterMock{}
	l := logger{title: "title", w: &w}
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

	l := logger{level: WARNING, w: &w}
	l.Warning("msg")
	l.Debug("msg")
	l.Error("msg")
	l.level = ERROR
	l.Error("msg")
	l.Debug("msg")
	l.Info("msg")
	l.level = INFO
	l.Info("msg")

	w.AssertNumberOfCalls(t, "Write", 4)
}
