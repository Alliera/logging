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
	logger := NewDefault("")

	if logger.w != os.Stdout {
		t.Errorf("Output direction must be Stdout")
	}
	if logger.level != WARNING && logger.level != DEBUG {
		t.Errorf("Log level must be WARNING or DEBUG, has: %v", logger.level)
	}
}

func TestSetFlags(t *testing.T) {
	logger := logger{}
	logger.SetFlags(Date | Date)
	assert.Equal(t, Date, logger.flag)
	logger.SetFlags(Date | Time)
	assert.Equal(t, Date|Time, logger.flag)
	logger.SetFlags(ShortCaller)
	assert.Equal(t, Date|Time|ShortCaller, logger.flag)
}

func TestUnsetFlags(t *testing.T) {
	logger := logger{}
	logger.SetFlags(Date | Time | Labels | Caller | ShortCaller)
	logger.UnsetFlags(Time)
	assert.Equal(t, Date|Labels|Caller|ShortCaller, logger.flag)
	logger.UnsetFlags(Labels | Caller)
	assert.Equal(t, Date|ShortCaller, logger.flag)
	logger.UnsetFlags(Labels)
	assert.Equal(t, Date|ShortCaller, logger.flag)
}

func TestIsLevelHigherThanDefault(t *testing.T) {
	logger := logger{}
	logger.level = FATAL
	assert.False(t, logger.isLevelHigherThanDefault(DEBUG))
	logger.level = DEBUG
	assert.True(t, logger.isLevelHigherThanDefault(DEBUG))
	logger.level = DEBUG
	assert.True(t, logger.isLevelHigherThanDefault(FATAL))
}

func TestLogNoFlags(t *testing.T) {
	w := WriterMock{}
	msg := "this is info message"
	w.On("Write", []byte(fmt.Sprintf("(title)  [INFO]  %s\n", msg)))
	logger := logger{title: "title", w: &w}
	logger.log(INFO, msg)
	w.AssertExpectations(t)
}

func TestLogWithFlags(t *testing.T) {
	w := WriterMock{}
	logger := logger{title: "title", w: &w}
	logger.SetFlags(Date | Time | Labels)

	msg := "this is info message"
	now := time.Now()

	w.On("Write", []byte(fmt.Sprintf(
		"DATE = %v  TIME =  %v  TITLE = (title)  "+
			"LEVEL = [INFO]  MSG = %s\n",
		now.Format("2006-01-02"),
		now.Format("15:04:05"),
		msg,
	)))

	logger.Info(msg)
	w.AssertExpectations(t)
}

func TestDifferentLogLevels(t *testing.T) {
	w := WriterMock{}

	w.On("Write", []byte("[WARNING]  msg\n"))
	w.On("Write", []byte("[DEBUG]  msg\n"))
	w.On("Write", []byte("[ERROR]  msg\n"))
	w.On("Write", []byte("[INFO]  msg\n"))

	logger := logger{level: WARNING, w: &w}
	logger.Warning("msg")
	logger.Debug("msg")
	logger.Error("msg")
	logger.level = ERROR
	logger.Error("msg")
	logger.Debug("msg")
	logger.Info("msg")
	logger.level = INFO
	logger.Info("msg")

	w.AssertNumberOfCalls(t, "Write", 4)
}
