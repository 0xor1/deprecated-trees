package misc

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/uber-go/zap"
	"runtime"
	"testing"
)

func Test_NewLog(t *testing.T) {
	l := NewLog(nil)
	assert.IsType(t, &log{}, l)
}

func Test_log_Location(t *testing.T) {
	logger := &mockLogger{}
	l := NewLog(logger)
	pc, file, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	logger.On("Log", zap.DebugLevel, "", zap.String("func", f.Name()), zap.String("file", file), zap.Int("line", line+4)).Return()

	l.Location()
	logger.AssertExpectations(t)
}

func Test_log_Debug(t *testing.T) {
	logger := &mockLogger{}
	l := NewLog(logger)
	pc, file, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	extraField := zap.Int("test", 0)
	logger.On("Log", zap.DebugLevel, "", zap.String("func", f.Name()), zap.String("file", file), zap.Int("line", line+5), extraField).Return()

	l.Debug(extraField)
	logger.AssertExpectations(t)
}

func Test_log_DebugErr(t *testing.T) {
	logger := &mockLogger{}
	l := NewLog(logger)
	pc, file, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	expectedErr := errors.New("test")
	logger.On("Log", zap.DebugLevel, "", zap.String("func", f.Name()), zap.String("file", file), zap.Int("line", line+5), zap.Error(expectedErr)).Return()

	err := l.DebugErr(expectedErr)
	logger.AssertExpectations(t)
	assert.Equal(t, expectedErr, err)
}

func Test_log_Info(t *testing.T) {
	logger := &mockLogger{}
	l := NewLog(logger)
	pc, file, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	extraField := zap.Int("test", 0)
	logger.On("Log", zap.InfoLevel, "", zap.String("func", f.Name()), zap.String("file", file), zap.Int("line", line+5), extraField).Return()

	l.Info(extraField)
	logger.AssertExpectations(t)
}

func Test_log_InfoErr(t *testing.T) {
	logger := &mockLogger{}
	l := NewLog(logger)
	pc, file, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	expectedErr := errors.New("test")
	logger.On("Log", zap.InfoLevel, "", zap.String("func", f.Name()), zap.String("file", file), zap.Int("line", line+5), zap.Error(expectedErr)).Return()

	err := l.InfoErr(expectedErr)
	logger.AssertExpectations(t)
	assert.Equal(t, expectedErr, err)
}

func Test_log_Warn(t *testing.T) {
	logger := &mockLogger{}
	l := NewLog(logger)
	pc, file, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	extraField := zap.Int("test", 0)
	logger.On("Log", zap.WarnLevel, "", zap.String("func", f.Name()), zap.String("file", file), zap.Int("line", line+5), extraField).Return()

	l.Warn(extraField)
	logger.AssertExpectations(t)
}

func Test_log_WarnErr(t *testing.T) {
	logger := &mockLogger{}
	l := NewLog(logger)
	pc, file, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	expectedErr := errors.New("test")
	logger.On("Log", zap.WarnLevel, "", zap.String("func", f.Name()), zap.String("file", file), zap.Int("line", line+5), zap.Error(expectedErr)).Return()

	err := l.WarnErr(expectedErr)
	logger.AssertExpectations(t)
	assert.Equal(t, expectedErr, err)
}

func Test_log_Error(t *testing.T) {
	logger := &mockLogger{}
	l := NewLog(logger)
	pc, file, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	extraField := zap.Int("test", 0)
	logger.On("Log", zap.ErrorLevel, "", zap.String("func", f.Name()), zap.String("file", file), zap.Int("line", line+5), extraField).Return()

	l.Error(extraField)
	logger.AssertExpectations(t)
}

func Test_log_ErrorErr(t *testing.T) {
	logger := &mockLogger{}
	l := NewLog(logger)
	pc, file, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	expectedErr := errors.New("test")
	logger.On("Log", zap.ErrorLevel, "", zap.String("func", f.Name()), zap.String("file", file), zap.Int("line", line+5), zap.Error(expectedErr)).Return()

	err := l.ErrorErr(expectedErr)
	logger.AssertExpectations(t)
	assert.Equal(t, expectedErr, err)
}

func Test_log_Panic(t *testing.T) {
	logger := &mockLogger{}
	l := NewLog(logger)
	pc, file, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	extraField := zap.Int("test", 0)
	logger.On("Log", zap.PanicLevel, "", zap.String("func", f.Name()), zap.String("file", file), zap.Int("line", line+5), extraField).Return()

	l.Panic(extraField)
	logger.AssertExpectations(t)
}

func Test_log_PanicErr(t *testing.T) {
	logger := &mockLogger{}
	l := NewLog(logger)
	pc, file, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	expectedErr := errors.New("test")
	logger.On("Log", zap.PanicLevel, "", zap.String("func", f.Name()), zap.String("file", file), zap.Int("line", line+5), zap.Error(expectedErr)).Return()

	err := l.PanicErr(expectedErr)
	logger.AssertExpectations(t)
	assert.Equal(t, expectedErr, err)
}

func Test_log_Fatal(t *testing.T) {
	logger := &mockLogger{}
	l := NewLog(logger)
	pc, file, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	extraField := zap.Int("test", 0)
	logger.On("Log", zap.FatalLevel, "", zap.String("func", f.Name()), zap.String("file", file), zap.Int("line", line+5), extraField).Return()

	l.Fatal(extraField)
	logger.AssertExpectations(t)
}

func Test_log_FatalErr(t *testing.T) {
	logger := &mockLogger{}
	l := NewLog(logger)
	pc, file, line, _ := runtime.Caller(0)
	f := runtime.FuncForPC(pc)
	expectedErr := errors.New("test")
	logger.On("Log", zap.FatalLevel, "", zap.String("func", f.Name()), zap.String("file", file), zap.Int("line", line+5), zap.Error(expectedErr)).Return()

	err := l.FatalErr(expectedErr)
	logger.AssertExpectations(t)
	assert.Equal(t, expectedErr, err)
}

//mocks

type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Log(level zap.Level, msg string, fields ...zap.Field) {
	args := make([]interface{}, 0, len(fields)+2)
	args = append(args, level, msg)
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}
